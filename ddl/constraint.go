// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package ddl

import (
	"fmt"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/failpoint"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/model"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/infoschema"
	"github.com/pingcap/tidb/meta"
	"github.com/pingcap/tidb/sessionctx"
	"github.com/pingcap/tidb/util/sqlexec"
)

func (w *worker) onAddCheckConstraint(d *ddlCtx, t *meta.Meta, job *model.Job) (ver int64, err error) {
	// Handle the rolling back job.
	if job.IsRollingback() {
		ver, err = onDropCheckConstraint(t, job)
		if err != nil {
			return ver, errors.Trace(err)
		}
		return ver, nil
	}

	failpoint.Inject("errorBeforeDecodeArgs", func(val failpoint.Value) {
		if val.(bool) {
			failpoint.Return(ver, errors.New("occur an error before decode args"))
		}
	})

	dbInfo, tblInfo, constraintInfoInMeta, constraintInfoInJob, err := checkAddCheckConstraint(t, job)
	if err != nil {
		return ver, errors.Trace(err)
	}
	if constraintInfoInMeta == nil {
		// It's first time to run add constraint job, so there is no constraint info in meta.
		// Use the raw constraint info from job directly and modify table info here.
		constraintInfoInJob.ID = allocateConstraintID(tblInfo)
		// Reset constraint name according to real-time constraints name at this point.
		constrNames := map[string]bool{}
		for _, constr := range tblInfo.Constraints {
			constrNames[constr.Name.L] = true
		}
		setNameForConstraintInfo(tblInfo.Name.L, constrNames, []*model.ConstraintInfo{constraintInfoInJob})
		// Double check the constraint dependency.
		existedColsMap := make(map[string]struct{})
		cols := tblInfo.Columns
		for _, v := range cols {
			if v.State == model.StatePublic {
				existedColsMap[v.Name.L] = struct{}{}
			}
		}
		dependedCols := constraintInfoInJob.ConstraintCols
		for _, k := range dependedCols {
			if _, ok := existedColsMap[k.L]; !ok {
				// The table constraint depended on a non-existed column.
				return ver, ErrTableCheckConstraintReferUnknown.GenWithStackByArgs(constraintInfoInJob.Name, k)
			}
		}

		tblInfo.Constraints = append(tblInfo.Constraints, constraintInfoInJob)
		constraintInfoInMeta = constraintInfoInJob
	}

	originalState := constraintInfoInMeta.State
	switch constraintInfoInMeta.State {
	case model.StateNone:
		// none -> write only
		job.SchemaState = model.StateWriteOnly
		constraintInfoInMeta.State = model.StateWriteOnly
		ver, err = updateVersionAndTableInfoWithCheck(t, job, tblInfo, originalState != constraintInfoInMeta.State)
	case model.StateWriteOnly:
		// write only -> public
		skipCheck := false
		failpoint.Inject("mockPassAddConstraintCheck", func(val failpoint.Value) {
			if val.(bool) {
				skipCheck = true
			}
		})
		if !skipCheck {
			err = w.addTableCheckConstraint(dbInfo, tblInfo, constraintInfoInMeta, job)
			if err != nil {
				return ver, errors.Trace(err)
			}
		}
		constraintInfoInMeta.State = model.StatePublic
		ver, err = updateVersionAndTableInfo(t, job, tblInfo, originalState != constraintInfoInMeta.State)
		if err != nil {
			return ver, errors.Trace(err)
		}

		// Finish this job.
		job.FinishTableJob(model.JobStateDone, model.StatePublic, ver, tblInfo)
	default:
		err = ErrInvalidDDLState.GenWithStackByArgs("constraint", constraintInfoInMeta.State)
	}

	return ver, errors.Trace(err)
}

// onDropCheckConstraint can be called from two case:
// 1: rollback in add constraint.(in rollback function the job.args will be changed)
// 2: user drop constraint ddl.
func onDropCheckConstraint(t *meta.Meta, job *model.Job) (ver int64, _ error) {
	tblInfo, constraintInfo, err := checkDropCheckConstraint(t, job)
	if err != nil {
		return ver, errors.Trace(err)
	}

	originalState := constraintInfo.State
	switch constraintInfo.State {
	case model.StatePublic:
		// public -> write only
		job.SchemaState = model.StateWriteOnly
		constraintInfo.State = model.StateWriteOnly
		ver, err = updateVersionAndTableInfoWithCheck(t, job, tblInfo, originalState != constraintInfo.State)
	case model.StateWriteOnly:
		// write only -> None
		// write only state constraint will still take effect to check the newly inserted data.
		// So the depended column shouldn't be dropped even in this intermediate state.
		constraintInfo.State = model.StateNone
		// remove the constraint from tableInfo.
		for i, constr := range tblInfo.Constraints {
			if constr.Name.L == constraintInfo.Name.L {
				tblInfo.Constraints = append(tblInfo.Constraints[0:i], tblInfo.Constraints[i+1:]...)
			}
		}
		ver, err = updateVersionAndTableInfo(t, job, tblInfo, originalState != constraintInfo.State)
		if err != nil {
			return ver, errors.Trace(err)
		}

		// Finish this job.
		if job.IsRollingback() {
			job.FinishTableJob(model.JobStateRollbackDone, model.StateNone, ver, tblInfo)
		} else {
			job.FinishTableJob(model.JobStateDone, model.StateNone, ver, tblInfo)
		}
	default:
		err = errInvalidDDLJob.GenWithStackByArgs("constraint", tblInfo.State)
	}
	return ver, errors.Trace(err)
}

func (w *worker) onAlterCheckConstraint(t *meta.Meta, job *model.Job) (ver int64, _ error) {
	dbInfo, tblInfo, constraintInfo, enforced, err := checkAlterCheckConstraint(t, job)
	if err != nil {
		return ver, errors.Trace(err)
	}

	// enforced will fetch table data and check the constraint.
	if constraintInfo.Enforced != enforced && enforced {
		skipCheck := false
		failpoint.Inject("mockPassAlterConstraintCheck", func(val failpoint.Value) {
			if val.(bool) {
				skipCheck = true
			}
		})
		if !skipCheck {
			err = w.addTableCheckConstraint(dbInfo, tblInfo, constraintInfo, job)
			if err != nil {
				// check constraint error will cancel the job, job state has been changed
				// to cancelled in addTableCheckConstraint.
				return ver, errors.Trace(err)
			}
		}
	}
	constraintInfo.Enforced = enforced
	ver, err = updateVersionAndTableInfoWithCheck(t, job, tblInfo, true)
	if err != nil {
		// update version and tableInfo error will cause retry.
		return ver, errors.Trace(err)
	}

	job.FinishTableJob(model.JobStateDone, model.StatePublic, ver, tblInfo)
	return ver, nil
}

func checkDropCheckConstraint(t *meta.Meta, job *model.Job) (*model.TableInfo, *model.ConstraintInfo, error) {
	schemaID := job.SchemaID
	tblInfo, err := getTableInfoAndCancelFaultJob(t, job, schemaID)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	var constrName model.CIStr
	err = job.DecodeArgs(&constrName)
	if err != nil {
		job.State = model.JobStateCancelled
		return nil, nil, errors.Trace(err)
	}

	// do the double-check with constraint existence.
	constraintInfo := tblInfo.FindConstraintInfoByName(constrName.L)
	if constraintInfo == nil {
		job.State = model.JobStateCancelled
		return nil, nil, ErrConstraintDoesNotExist.GenWithStackByArgs(constrName)
	}
	return tblInfo, constraintInfo, nil
}

func checkAddCheckConstraint(t *meta.Meta, job *model.Job) (*model.DBInfo, *model.TableInfo, *model.ConstraintInfo, *model.ConstraintInfo, error) {
	schemaID := job.SchemaID
	dbInfo, err := t.GetDatabase(job.SchemaID)
	if err != nil {
		return nil, nil, nil, nil, errors.Trace(err)
	}
	tblInfo, err := getTableInfoAndCancelFaultJob(t, job, schemaID)
	if err != nil {
		return nil, nil, nil, nil, errors.Trace(err)
	}
	constraintInfo1 := &model.ConstraintInfo{}
	err = job.DecodeArgs(constraintInfo1)
	if err != nil {
		job.State = model.JobStateCancelled
		return nil, nil, nil, nil, errors.Trace(err)
	}
	// do the double-check with constraint existence.
	constraintInfo2 := tblInfo.FindConstraintInfoByName(constraintInfo1.Name.L)
	if constraintInfo2 != nil {
		if constraintInfo2.State == model.StatePublic {
			// We already have a constraint with the same constraint name.
			job.State = model.JobStateCancelled
			return nil, nil, nil, nil, infoschema.ErrColumnExists.GenWithStackByArgs(constraintInfo1.Name)
		}
		// if not, that means constraint was in intermediate state.
	}
	return dbInfo, tblInfo, constraintInfo2, constraintInfo1, nil
}

func checkAlterCheckConstraint(t *meta.Meta, job *model.Job) (*model.DBInfo, *model.TableInfo, *model.ConstraintInfo, bool, error) {
	schemaID := job.SchemaID
	dbInfo, err := t.GetDatabase(job.SchemaID)
	if err != nil {
		return nil, nil, nil, false, errors.Trace(err)
	}
	tblInfo, err := getTableInfoAndCancelFaultJob(t, job, schemaID)
	if err != nil {
		return nil, nil, nil, false, errors.Trace(err)
	}

	var (
		enforced   bool
		constrName model.CIStr
	)
	err = job.DecodeArgs(&constrName, &enforced)
	if err != nil {
		job.State = model.JobStateCancelled
		return nil, nil, nil, false, errors.Trace(err)
	}

	// do the double-check with constraint existence.
	constraintInfo := tblInfo.FindConstraintInfoByName(constrName.L)
	if constraintInfo == nil {
		job.State = model.JobStateCancelled
		return nil, nil, nil, false, ErrConstraintDoesNotExist.GenWithStackByArgs(constrName)
	}
	return dbInfo, tblInfo, constraintInfo, enforced, nil
}

func allocateConstraintID(tblInfo *model.TableInfo) int64 {
	tblInfo.MaxConstraintID++
	return tblInfo.MaxConstraintID
}

func buildConstraintInfo(tblInfo *model.TableInfo, dependedCols []model.CIStr, constr *ast.Constraint, state model.SchemaState) (*model.ConstraintInfo, error) {
	constraintName := model.NewCIStr(constr.Name)
	if err := checkTooLongConstraint(constraintName); err != nil {
		return nil, errors.Trace(err)
	}

	// Restore check constraint expression to string.
	var sb strings.Builder
	restoreFlags := format.RestoreStringSingleQuotes | format.RestoreKeyWordLowercase | format.RestoreNameBackQuotes |
		format.RestoreSpacesAroundBinaryOperation
	restoreCtx := format.NewRestoreCtx(restoreFlags, &sb)

	sb.Reset()
	err := constr.Expr.Restore(restoreCtx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Create constraint info.
	constraintInfo := &model.ConstraintInfo{
		Name:           constraintName,
		Table:          tblInfo.Name,
		ConstraintCols: dependedCols,
		ExprString:     sb.String(),
		Enforced:       constr.Enforced,
		InColumn:       constr.InColumn,
		State:          state,
	}

	return constraintInfo, nil
}

func checkTooLongConstraint(constr model.CIStr) error {
	if len(constr.L) > mysql.MaxConstraintIdentifierLen {
		return ErrTooLongIdent.GenWithStackByArgs(constr)
	}
	return nil
}

// findDependedColsMapInExpr returns a set of string, which indicates
// the names of the columns that are depended by exprNode.
func findDependedColsMapInExpr(expr ast.ExprNode) map[string]struct{} {
	colNames := findColumnNamesInExpr(expr)
	colsMap := make(map[string]struct{}, len(colNames))
	for _, depCol := range colNames {
		colsMap[depCol.Name.L] = struct{}{}
	}
	return colsMap
}

func (w *worker) addTableCheckConstraint(dbInfo *model.DBInfo, tableInfo *model.TableInfo, constr *model.ConstraintInfo, job *model.Job) error {
	// Get sessionctx from ddl context resource pool in ddl worker.
	var ctx sessionctx.Context
	ctx, err := w.sessPool.get()
	if err != nil {
		return errors.Trace(err)
	}
	defer w.sessPool.put(ctx)

	// If there is any row can't pass the check expression, the add constraint action will error.
	// It's no need to construct expression node out and pull the chunk rows through it. Here we
	// can let the check expression restored string as the filter in where clause directly.
	// Prepare internal SQL to fetch data from physical table under this filter.
	sql := fmt.Sprintf("select count(1) from `%s`.`%s` where not %s limit 1", dbInfo.Name.L, tableInfo.Name.L, constr.ExprString)
	rows, _, err := ctx.(sqlexec.RestrictedSQLExecutor).ExecRestrictedSQL(sql)
	if err != nil {
		return errors.Trace(err)
	}
	rowCount := rows[0].GetInt64(0)
	if rowCount != 0 {
		// If check constraint fail, the job state should be changed to canceled, otherwise it will tracked in.
		job.State = model.JobStateCancelled
		return ErrCheckConstraintIsViolated.GenWithStackByArgs(constr.Name.L)
	}
	return nil
}

func setNameForConstraintInfo(tableLowerName string, namesMap map[string]bool, infos []*model.ConstraintInfo) {
	cnt := 1
	constraintPrefix := tableLowerName + "_chk_"
	for _, constrInfo := range infos {
		if constrInfo.Name.O == "" {
			constrName := fmt.Sprintf("%s%d", constraintPrefix, cnt)
			for {
				// loop until find constrName that haven't been used.
				if !namesMap[constrName] {
					namesMap[constrName] = true
					break
				}
				cnt++
				constrName = fmt.Sprintf("%s%d", constraintPrefix, cnt)
			}
			constrInfo.Name = model.NewCIStr(constrName)
		}
	}
}
