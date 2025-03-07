// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expression

import (
	"fmt"
	"math"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/opcode"
	"github.com/pingcap/tidb/pkg/sessionctx"
	"github.com/pingcap/tidb/pkg/types"
	"github.com/pingcap/tidb/pkg/util/chunk"
	"github.com/pingcap/tipb/go-tipb"
)

var (
	_ functionClass = &logicAndFunctionClass{}
	_ functionClass = &logicOrFunctionClass{}
	_ functionClass = &logicXorFunctionClass{}
	_ functionClass = &isTrueOrFalseFunctionClass{}
	_ functionClass = &unaryMinusFunctionClass{}
	_ functionClass = &isNullFunctionClass{}
	_ functionClass = &unaryNotFunctionClass{}
)

var (
	_ builtinFunc = &builtinLogicAndSig{}
	_ builtinFunc = &builtinLogicOrSig{}
	_ builtinFunc = &builtinLogicXorSig{}
	_ builtinFunc = &builtinRealIsTrueSig{}
	_ builtinFunc = &builtinDecimalIsTrueSig{}
	_ builtinFunc = &builtinIntIsTrueSig{}
	_ builtinFunc = &builtinRealIsFalseSig{}
	_ builtinFunc = &builtinDecimalIsFalseSig{}
	_ builtinFunc = &builtinIntIsFalseSig{}
	_ builtinFunc = &builtinUnaryMinusIntSig{}
	_ builtinFunc = &builtinDecimalIsNullSig{}
	_ builtinFunc = &builtinDurationIsNullSig{}
	_ builtinFunc = &builtinIntIsNullSig{}
	_ builtinFunc = &builtinRealIsNullSig{}
	_ builtinFunc = &builtinStringIsNullSig{}
	_ builtinFunc = &builtinTimeIsNullSig{}
	_ builtinFunc = &builtinUnaryNotRealSig{}
	_ builtinFunc = &builtinUnaryNotDecimalSig{}
	_ builtinFunc = &builtinUnaryNotIntSig{}
)

type logicAndFunctionClass struct {
	baseFunctionClass
}

func (c *logicAndFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	args[0], err = wrapWithIsTrue(ctx, true, args[0], false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	args[1], err = wrapWithIsTrue(ctx, true, args[1], false)
	if err != nil {
		return nil, errors.Trace(err)
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinLogicAndSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_LogicalAnd)
	sig.tp.SetFlen(1)
	return sig, nil
}

type builtinLogicAndSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicAndSig) Clone() builtinFunc {
	newSig := &builtinLogicAndSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinLogicAndSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull0, err := b.args[0].EvalInt(ctx, row)
	if err != nil || (!isNull0 && arg0 == 0) {
		return 0, err != nil, err
	}
	arg1, isNull1, err := b.args[1].EvalInt(ctx, row)
	if err != nil || (!isNull1 && arg1 == 0) {
		return 0, err != nil, err
	}
	if isNull0 || isNull1 {
		return 0, true, nil
	}
	return 1, false, nil
}

type logicOrFunctionClass struct {
	baseFunctionClass
}

func (c *logicOrFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	args[0], err = wrapWithIsTrue(ctx, true, args[0], false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	args[1], err = wrapWithIsTrue(ctx, true, args[1], false)
	if err != nil {
		return nil, errors.Trace(err)
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	bf.tp.SetFlen(1)
	sig := &builtinLogicOrSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_LogicalOr)
	return sig, nil
}

type builtinLogicOrSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicOrSig) Clone() builtinFunc {
	newSig := &builtinLogicOrSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinLogicOrSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull0, err := b.args[0].EvalInt(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if !isNull0 && arg0 != 0 {
		return 1, false, nil
	}
	arg1, isNull1, err := b.args[1].EvalInt(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if !isNull1 && arg1 != 0 {
		return 1, false, nil
	}
	if isNull0 || isNull1 {
		return 0, true, nil
	}
	return 0, false, nil
}

type logicXorFunctionClass struct {
	baseFunctionClass
}

func (c *logicXorFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	args[0], err = wrapWithIsTrue(ctx, true, args[0], false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	args[1], err = wrapWithIsTrue(ctx, true, args[1], false)
	if err != nil {
		return nil, errors.Trace(err)
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinLogicXorSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_LogicalXor)
	sig.tp.SetFlen(1)
	return sig, nil
}

type builtinLogicXorSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicXorSig) Clone() builtinFunc {
	newSig := &builtinLogicXorSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinLogicXorSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	if (arg0 != 0 && arg1 != 0) || (arg0 == 0 && arg1 == 0) {
		return 0, false, nil
	}
	return 1, false, nil
}

type bitAndFunctionClass struct {
	baseFunctionClass
}

func (c *bitAndFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinBitAndSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_BitAndSig)
	sig.tp.AddFlag(mysql.UnsignedFlag)
	return sig, nil
}

type builtinBitAndSig struct {
	baseBuiltinFunc
}

func (b *builtinBitAndSig) Clone() builtinFunc {
	newSig := &builtinBitAndSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinBitAndSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return arg0 & arg1, false, nil
}

type bitOrFunctionClass struct {
	baseFunctionClass
}

func (c *bitOrFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinBitOrSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_BitOrSig)
	sig.tp.AddFlag(mysql.UnsignedFlag)
	return sig, nil
}

type builtinBitOrSig struct {
	baseBuiltinFunc
}

func (b *builtinBitOrSig) Clone() builtinFunc {
	newSig := &builtinBitOrSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinBitOrSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return arg0 | arg1, false, nil
}

type bitXorFunctionClass struct {
	baseFunctionClass
}

func (c *bitXorFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinBitXorSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_BitXorSig)
	sig.tp.AddFlag(mysql.UnsignedFlag)
	return sig, nil
}

type builtinBitXorSig struct {
	baseBuiltinFunc
}

func (b *builtinBitXorSig) Clone() builtinFunc {
	newSig := &builtinBitXorSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinBitXorSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return arg0 ^ arg1, false, nil
}

type leftShiftFunctionClass struct {
	baseFunctionClass
}

func (c *leftShiftFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinLeftShiftSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_LeftShift)
	sig.tp.AddFlag(mysql.UnsignedFlag)
	return sig, nil
}

type builtinLeftShiftSig struct {
	baseBuiltinFunc
}

func (b *builtinLeftShiftSig) Clone() builtinFunc {
	newSig := &builtinLeftShiftSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinLeftShiftSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return int64(uint64(arg0) << uint64(arg1)), false, nil
}

type rightShiftFunctionClass struct {
	baseFunctionClass
}

func (c *rightShiftFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	sig := &builtinRightShiftSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_RightShift)
	sig.tp.AddFlag(mysql.UnsignedFlag)
	return sig, nil
}

type builtinRightShiftSig struct {
	baseBuiltinFunc
}

func (b *builtinRightShiftSig) Clone() builtinFunc {
	newSig := &builtinRightShiftSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinRightShiftSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg0, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	arg1, isNull, err := b.args[1].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return int64(uint64(arg0) >> uint64(arg1)), false, nil
}

type isTrueOrFalseFunctionClass struct {
	baseFunctionClass
	op opcode.Op

	// keepNull indicates how this function treats a null input parameter.
	// If keepNull is true and the input parameter is null, the function will return null.
	// If keepNull is false, the null input parameter will be cast to 0.
	keepNull bool
}

func (c *isTrueOrFalseFunctionClass) getDisplayName() string {
	var nameBuilder strings.Builder
	c.op.Format(&nameBuilder)
	return nameBuilder.String()
}

func (c *isTrueOrFalseFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}

	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp || argTp == types.ETDatetime || argTp == types.ETDuration || argTp == types.ETJson || argTp == types.ETString {
		argTp = types.ETReal
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, argTp)
	if err != nil {
		return nil, err
	}
	bf.tp.SetFlen(1)

	var sig builtinFunc
	switch c.op {
	case opcode.IsTruth:
		switch argTp {
		case types.ETReal:
			sig = &builtinRealIsTrueSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_RealIsTrueWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_RealIsTrue)
			}
		case types.ETDecimal:
			sig = &builtinDecimalIsTrueSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_DecimalIsTrueWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_DecimalIsTrue)
			}
		case types.ETInt:
			sig = &builtinIntIsTrueSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_IntIsTrueWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_IntIsTrue)
			}
		default:
			return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
		}
	case opcode.IsFalsity:
		switch argTp {
		case types.ETReal:
			sig = &builtinRealIsFalseSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_RealIsFalseWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_RealIsFalse)
			}
		case types.ETDecimal:
			sig = &builtinDecimalIsFalseSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_DecimalIsFalseWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_DecimalIsFalse)
			}
		case types.ETInt:
			sig = &builtinIntIsFalseSig{bf, c.keepNull}
			if c.keepNull {
				sig.setPbCode(tipb.ScalarFuncSig_IntIsFalseWithNull)
			} else {
				sig.setPbCode(tipb.ScalarFuncSig_IntIsFalse)
			}
		default:
			return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
		}
	}
	return sig, nil
}

type builtinRealIsTrueSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinRealIsTrueSig) Clone() builtinFunc {
	newSig := &builtinRealIsTrueSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinRealIsTrueSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalReal(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || input == 0 {
		return 0, false, nil
	}
	return 1, false, nil
}

type builtinDecimalIsTrueSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinDecimalIsTrueSig) Clone() builtinFunc {
	newSig := &builtinDecimalIsTrueSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinDecimalIsTrueSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalDecimal(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || input.IsZero() {
		return 0, false, nil
	}
	return 1, false, nil
}

type builtinIntIsTrueSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinIntIsTrueSig) Clone() builtinFunc {
	newSig := &builtinIntIsTrueSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinIntIsTrueSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalInt(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || input == 0 {
		return 0, false, nil
	}
	return 1, false, nil
}

type builtinRealIsFalseSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinRealIsFalseSig) Clone() builtinFunc {
	newSig := &builtinRealIsFalseSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinRealIsFalseSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalReal(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || input != 0 {
		return 0, false, nil
	}
	return 1, false, nil
}

type builtinDecimalIsFalseSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinDecimalIsFalseSig) Clone() builtinFunc {
	newSig := &builtinDecimalIsFalseSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinDecimalIsFalseSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalDecimal(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || !input.IsZero() {
		return 0, false, nil
	}
	return 1, false, nil
}

type builtinIntIsFalseSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinIntIsFalseSig) Clone() builtinFunc {
	newSig := &builtinIntIsFalseSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinIntIsFalseSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	input, isNull, err := b.args[0].EvalInt(ctx, row)
	if err != nil {
		return 0, true, err
	}
	if b.keepNull && isNull {
		return 0, true, nil
	}
	if isNull || input != 0 {
		return 0, false, nil
	}
	return 1, false, nil
}

type bitNegFunctionClass struct {
	baseFunctionClass
}

func (c *bitNegFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt)
	if err != nil {
		return nil, err
	}
	bf.tp.AddFlag(mysql.UnsignedFlag)
	sig := &builtinBitNegSig{bf}
	sig.setPbCode(tipb.ScalarFuncSig_BitNegSig)
	return sig, nil
}

type builtinBitNegSig struct {
	baseBuiltinFunc
}

func (b *builtinBitNegSig) Clone() builtinFunc {
	newSig := &builtinBitNegSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinBitNegSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, isNull, err
	}
	return ^arg, false, nil
}

type unaryNotFunctionClass struct {
	baseFunctionClass
}

func (c *unaryNotFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}

	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp || argTp == types.ETDatetime || argTp == types.ETDuration {
		argTp = types.ETInt
	} else if argTp == types.ETString {
		argTp = types.ETReal
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, argTp)
	if err != nil {
		return nil, err
	}
	bf.tp.SetFlen(1)

	var sig builtinFunc
	switch argTp {
	case types.ETReal:
		sig = &builtinUnaryNotRealSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryNotReal)
	case types.ETDecimal:
		sig = &builtinUnaryNotDecimalSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryNotDecimal)
	case types.ETInt:
		sig = &builtinUnaryNotIntSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryNotInt)
	case types.ETJson:
		ctx.GetSessionVars().StmtCtx.AppendWarning(errJSONInBooleanContext)
		sig = &builtinUnaryNotJSONSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryNotJSON)
	default:
		return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
	}
	return sig, nil
}

type builtinUnaryNotRealSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotRealSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotRealSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotRealSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalReal(ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg == 0 {
		return 1, false, nil
	}
	return 0, false, nil
}

type builtinUnaryNotDecimalSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotDecimalSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotDecimalSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotDecimalSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalDecimal(ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg.IsZero() {
		return 1, false, nil
	}
	return 0, false, nil
}

type builtinUnaryNotIntSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotIntSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotIntSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotIntSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalInt(ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg == 0 {
		return 1, false, nil
	}
	return 0, false, nil
}

type builtinUnaryNotJSONSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotJSONSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotJSONSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotJSONSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalJSON(ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}

	if types.CompareBinaryJSON(arg, types.CreateBinaryJSON(int64(0))) == 0 {
		return 1, false, nil
	}
	return 0, false, nil
}

type unaryMinusFunctionClass struct {
	baseFunctionClass
}

func (c *unaryMinusFunctionClass) handleIntOverflow(arg *Constant) (overflow bool) {
	if mysql.HasUnsignedFlag(arg.GetType().GetFlag()) {
		uval := arg.Value.GetUint64()
		// -math.MinInt64 is 9223372036854775808, so if uval is more than 9223372036854775808, like
		// 9223372036854775809, -9223372036854775809 is less than math.MinInt64, overflow occurs.
		if uval > uint64(-math.MinInt64) {
			return true
		}
	} else {
		val := arg.Value.GetInt64()
		// The math.MinInt64 is -9223372036854775808, the math.MaxInt64 is 9223372036854775807,
		// which is less than abs(-9223372036854775808). When val == math.MinInt64, overflow occurs.
		if val == math.MinInt64 {
			return true
		}
	}
	return false
}

// typeInfer infers unaryMinus function return type. when the arg is an int constant and overflow,
// typerInfer will infers the return type as types.ETDecimal, not types.ETInt.
func (c *unaryMinusFunctionClass) typeInfer(argExpr Expression) (types.EvalType, bool) {
	tp := argExpr.GetType().EvalType()
	if tp != types.ETInt && tp != types.ETDecimal {
		tp = types.ETReal
	}

	overflow := false
	// TODO: Handle float overflow.
	if arg, ok := argExpr.(*Constant); ok && tp == types.ETInt {
		overflow = c.handleIntOverflow(arg)
		if overflow {
			tp = types.ETDecimal
		}
	}
	return tp, overflow
}

func (c *unaryMinusFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (sig builtinFunc, err error) {
	if err = c.verifyArgs(args); err != nil {
		return nil, err
	}

	argExpr, argExprTp := args[0], args[0].GetType()
	_, intOverflow := c.typeInfer(argExpr)

	var bf baseBuiltinFunc
	switch argExprTp.EvalType() {
	case types.ETInt:
		if intOverflow {
			bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETDecimal, types.ETDecimal)
			if err != nil {
				return nil, err
			}
			sig = &builtinUnaryMinusDecimalSig{bf, true}
			sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
		} else {
			bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, types.ETInt)
			if err != nil {
				return nil, err
			}
			sig = &builtinUnaryMinusIntSig{bf}
			sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusInt)
		}
		bf.tp.SetDecimal(0)
	case types.ETDecimal:
		bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETDecimal, types.ETDecimal)
		if err != nil {
			return nil, err
		}
		bf.tp.SetDecimalUnderLimit(argExprTp.GetDecimal())
		sig = &builtinUnaryMinusDecimalSig{bf, false}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
	case types.ETReal:
		bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETReal, types.ETReal)
		if err != nil {
			return nil, err
		}
		sig = &builtinUnaryMinusRealSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusReal)
	default:
		tp := argExpr.GetType().GetType()
		if types.IsTypeTime(tp) || tp == mysql.TypeDuration {
			bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETDecimal, types.ETDecimal)
			if err != nil {
				return nil, err
			}
			sig = &builtinUnaryMinusDecimalSig{bf, false}
			sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
		} else {
			bf, err = newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETReal, types.ETReal)
			if err != nil {
				return nil, err
			}
			sig = &builtinUnaryMinusRealSig{bf}
			sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusReal)
		}
	}
	bf.tp.SetFlenUnderLimit(argExprTp.GetFlen() + 1)
	return sig, err
}

type builtinUnaryMinusIntSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryMinusIntSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusIntSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusIntSig) evalInt(ctx EvalContext, row chunk.Row) (res int64, isNull bool, err error) {
	var val int64
	val, isNull, err = b.args[0].EvalInt(ctx, row)
	if err != nil || isNull {
		return val, isNull, err
	}

	if mysql.HasUnsignedFlag(b.args[0].GetType().GetFlag()) {
		uval := uint64(val)
		if uval > uint64(-math.MinInt64) {
			return 0, false, types.ErrOverflow.GenWithStackByArgs("BIGINT", fmt.Sprintf("-%v", uval))
		} else if uval == uint64(-math.MinInt64) {
			return math.MinInt64, false, nil
		}
	} else if val == math.MinInt64 {
		return 0, false, types.ErrOverflow.GenWithStackByArgs("BIGINT", fmt.Sprintf("-%v", val))
	}
	return -val, false, nil
}

type builtinUnaryMinusDecimalSig struct {
	baseBuiltinFunc

	constantArgOverflow bool
}

func (b *builtinUnaryMinusDecimalSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusDecimalSig{constantArgOverflow: b.constantArgOverflow}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusDecimalSig) evalDecimal(ctx EvalContext, row chunk.Row) (*types.MyDecimal, bool, error) {
	dec, isNull, err := b.args[0].EvalDecimal(ctx, row)
	if err != nil || isNull {
		return dec, isNull, err
	}
	return types.DecimalNeg(dec), false, nil
}

type builtinUnaryMinusRealSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryMinusRealSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusRealSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusRealSig) evalReal(ctx EvalContext, row chunk.Row) (float64, bool, error) {
	val, isNull, err := b.args[0].EvalReal(ctx, row)
	return -val, isNull, err
}

type isNullFunctionClass struct {
	baseFunctionClass
}

func (c *isNullFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}
	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp {
		argTp = types.ETDatetime
	} else if argTp == types.ETJson {
		argTp = types.ETString
	}
	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETInt, argTp)
	if err != nil {
		return nil, err
	}
	bf.tp.SetFlen(1)
	var sig builtinFunc
	switch argTp {
	case types.ETInt:
		sig = &builtinIntIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_IntIsNull)
	case types.ETDecimal:
		sig = &builtinDecimalIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_DecimalIsNull)
	case types.ETReal:
		sig = &builtinRealIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_RealIsNull)
	case types.ETDatetime:
		sig = &builtinTimeIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_TimeIsNull)
	case types.ETDuration:
		sig = &builtinDurationIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_DurationIsNull)
	case types.ETString:
		sig = &builtinStringIsNullSig{bf}
		sig.setPbCode(tipb.ScalarFuncSig_StringIsNull)
	default:
		panic("unexpected types.EvalType")
	}
	return sig, nil
}

type builtinDecimalIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinDecimalIsNullSig) Clone() builtinFunc {
	newSig := &builtinDecimalIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func evalIsNull(isNull bool, err error) (int64, bool, error) {
	if err != nil {
		return 0, true, err
	}
	if isNull {
		return 1, false, nil
	}
	return 0, false, nil
}

func (b *builtinDecimalIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalDecimal(ctx, row)
	return evalIsNull(isNull, err)
}

type builtinDurationIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinDurationIsNullSig) Clone() builtinFunc {
	newSig := &builtinDurationIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinDurationIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalDuration(ctx, row)
	return evalIsNull(isNull, err)
}

type builtinIntIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinIntIsNullSig) Clone() builtinFunc {
	newSig := &builtinIntIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinIntIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalInt(ctx, row)
	return evalIsNull(isNull, err)
}

type builtinRealIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinRealIsNullSig) Clone() builtinFunc {
	newSig := &builtinRealIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinRealIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalReal(ctx, row)
	return evalIsNull(isNull, err)
}

type builtinStringIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinStringIsNullSig) Clone() builtinFunc {
	newSig := &builtinStringIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinStringIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalString(ctx, row)
	return evalIsNull(isNull, err)
}

type builtinTimeIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinTimeIsNullSig) Clone() builtinFunc {
	newSig := &builtinTimeIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinTimeIsNullSig) evalInt(ctx EvalContext, row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalTime(ctx, row)
	return evalIsNull(isNull, err)
}
