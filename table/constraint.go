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

package table

import (
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/model"
	"github.com/pingcap/tidb/expression"
	"github.com/pingcap/tidb/sessionctx"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/logutil"
	"github.com/pingcap/tidb/util/mock"
	"go.uber.org/zap"
)

// Constraint provides meta and map dependency describing a table constraint.
type Constraint struct {
	*model.ConstraintInfo

	ConstraintExpr expression.Expression
}

// ToConstraint ...
func ToConstraint(constraintInfo *model.ConstraintInfo, tblInfo *model.TableInfo) (*Constraint, error) {
	ctx := mock.NewContext()
	dbName := model.NewCIStr(ctx.GetSessionVars().CurrentDB)
	columns, names := expression.ColumnInfos2ColumnsAndNames(ctx, dbName, tblInfo.Name, tblInfo.Columns)
	expr, err := buildConstraintExpression(ctx, constraintInfo.ExprString, columns, names)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Constraint{
		constraintInfo,
		expr,
	}, nil
}

func buildConstraintExpression(ctx sessionctx.Context, exprString string,
	columns []*expression.Column, names types.NameSlice) (expression.Expression, error) {
	schema := expression.NewSchema(columns...)
	exprs, err := expression.ParseSimpleExprsWithNames(ctx, exprString, schema, names)
	if err != nil {
		// If it got an error here, ddl may hang forever, so this error log is important.
		logutil.BgLogger().Error("wrong check constraint expression", zap.String("expression", exprString), zap.Error(err))
		return nil, errors.Trace(err)
	}
	return exprs[0], nil
}

// ToInfo ...
func (c *Constraint) ToInfo() *model.ConstraintInfo {
	return c.ConstraintInfo
}
