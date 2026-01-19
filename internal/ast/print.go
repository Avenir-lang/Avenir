package ast

import (
	"fmt"
	"io"
	"strings"
)

// Dump returns a human-readable representation of the AST.
func Dump(node Node) string {
	var sb strings.Builder
	fprintNode(&sb, node, 0)
	return sb.String()
}

func fprintNode(w io.Writer, n Node, indent int) {
	if n == nil {
		return
	}

	ind := strings.Repeat("  ", indent)

	switch n := n.(type) {
	case *Program:
		fmt.Fprintf(w, "%sProgram\n", ind)
		if n.Package != nil {
			fprintNode(w, n.Package, indent+1)
		}
		if len(n.Imports) > 0 {
			fmt.Fprintf(w, "%s  Imports:\n", ind)
			for _, imp := range n.Imports {
				fprintNode(w, imp, indent+2)
			}
		}
		for _, fn := range n.Funcs {
			fprintNode(w, fn, indent+1)
		}

	case *PackageDecl:
		fmt.Fprintf(w, "%sPackageDecl name=%s\n", ind, n.Name)

	case *ImportDecl:
		pathStr := strings.Join(n.Path, ".")
		aliasStr := n.Alias
		if aliasStr == "" {
			aliasStr = "<default>"
		}
		fmt.Fprintf(w, "%sImportDecl path=%s alias=%s\n", ind, pathStr, aliasStr)

	case *FunDecl:
		pubStr := ""
		if n.IsPublic {
			pubStr = " pub"
		}
		methodStr := ""
		if n.Receiver != nil {
			// Get type name for display
			typeName := "unknown"
			if simpleType, ok := n.Receiver.Type.(*SimpleType); ok {
				typeName = simpleType.Name
			}
			if n.Receiver.Kind == ReceiverStatic {
				methodStr = fmt.Sprintf(" static method on %s", typeName)
			} else {
				methodStr = fmt.Sprintf(" instance method on %s (receiver: %s)", typeName, n.Receiver.Name)
			}
		}
		fmt.Fprintf(w, "%sFunDecl name=%s%s%s\n", ind, n.Name, pubStr, methodStr)
		if n.Receiver != nil {
			fmt.Fprintf(w, "%s  Receiver:\n", ind)
			if n.Receiver.Kind == ReceiverStatic {
				fmt.Fprintf(w, "%s    Kind: static\n", ind)
			} else {
				fmt.Fprintf(w, "%s    Kind: instance\n", ind)
				fmt.Fprintf(w, "%s    Name: %s\n", ind, n.Receiver.Name)
			}
			fmt.Fprintf(w, "%s    Type:\n", ind)
			fprintNode(w, n.Receiver.Type, indent+3)
		}
		if len(n.Params) > 0 {
			fmt.Fprintf(w, "%s  Params:\n", ind)
			for _, p := range n.Params {
				fprintNode(w, p, indent+2)
			}
		}
		if n.Return != nil {
			fmt.Fprintf(w, "%s  ReturnType:\n", ind)
			fprintNode(w, n.Return, indent+2)
		}
		if n.Body != nil {
			fmt.Fprintf(w, "%s  Body:\n", ind)
			fprintNode(w, n.Body, indent+2)
		}

	case *Param:
		fmt.Fprintf(w, "%sParam name=%s\n", ind, n.Name)
		fprintNode(w, n.Type, indent+1)
		if n.Default != nil {
			fmt.Fprintf(w, "%s  Default:\n", ind)
			fprintNode(w, n.Default, indent+2)
		}

	case *SimpleType:
		fmt.Fprintf(w, "%sSimpleType %s\n", ind, n.Name)

	case *QualifiedType:
		fmt.Fprintf(w, "%sQualifiedType %s\n", ind, strings.Join(n.Path, "."))

	case *ListType:
		fmt.Fprintf(w, "%sListType\n", ind)
		for _, t := range n.ElementTypes {
			fprintNode(w, t, indent+1)
		}

	case *DictType:
		fmt.Fprintf(w, "%sDictType\n", ind)
		fprintNode(w, n.ValueType, indent+1)

	case *UnionType:
		fmt.Fprintf(w, "%sUnionType\n", ind)
		for _, t := range n.Variants {
			fprintNode(w, t, indent+1)
		}

	case *OptionalType:
		fmt.Fprintf(w, "%sOptionalType\n", ind)
		fprintNode(w, n.Inner, indent+1)

	case *StructDecl:
		pubStr := ""
		if n.IsPublic {
			pubStr = " pub"
		}
		fmt.Fprintf(w, "%sStructDecl%s name=%s\n", ind, pubStr, n.Name)
		for _, f := range n.Fields {
			fmt.Fprintf(w, "%s  Field %s:\n", ind, f.Name)
			fprintNode(w, f.Type, indent+2)
		}

	case *FieldDecl:
		fmt.Fprintf(w, "%sFieldDecl name=%s\n", ind, n.Name)
		fprintNode(w, n.Type, indent+1)

	case *StructLiteral:
		fmt.Fprintf(w, "%sStructLiteral type=%s\n", ind, n.TypeName)
		for _, f := range n.Fields {
			fmt.Fprintf(w, "%s  Field %s:\n", ind, f.Name)
			fprintNode(w, f.Value, indent+2)
		}

	case *FieldInit:
		fmt.Fprintf(w, "%sFieldInit name=%s\n", ind, n.Name)
		fprintNode(w, n.Value, indent+1)

	case *BlockStmt:
		fmt.Fprintf(w, "%sBlockStmt\n", ind)
		for _, s := range n.Stmts {
			fprintNode(w, s, indent+1)
		}

	case *VarDeclStmt:
		fmt.Fprintf(w, "%sVarDecl name=%s\n", ind, n.Name)
		fmt.Fprintf(w, "%s  Type:\n", ind)
		fprintNode(w, n.Type, indent+2)
		fmt.Fprintf(w, "%s  Value:\n", ind)
		fprintNode(w, n.Value, indent+2)

	case *AssignStmt:
		fmt.Fprintf(w, "%sAssign name=%s\n", ind, n.Name)
		fprintNode(w, n.Value, indent+1)

	case *ExprStmt:
		fmt.Fprintf(w, "%sExprStmt\n", ind)
		fprintNode(w, n.Expression, indent+1)

	case *IfStmt:
		fmt.Fprintf(w, "%sIfStmt\n", ind)
		fmt.Fprintf(w, "%s  Cond:\n", ind)
		fprintNode(w, n.Cond, indent+2)
		fmt.Fprintf(w, "%s  Then:\n", ind)
		fprintNode(w, n.Then, indent+2)
		if n.Else != nil {
			fmt.Fprintf(w, "%s  Else:\n", ind)
			fprintNode(w, n.Else.(Node), indent+2)
		}

	case *ReturnStmt:
		fmt.Fprintf(w, "%sReturnStmt\n", ind)
		if n.Result != nil {
			fprintNode(w, n.Result, indent+1)
		}

	case *WhileStmt:
		fmt.Fprintf(w, "%sWhileStmt\n", ind)
		fmt.Fprintf(w, "%s  Cond:\n", ind)
		fprintNode(w, n.Cond, indent+2)
		fmt.Fprintf(w, "%s  Body:\n", ind)
		fprintNode(w, n.Body, indent+2)

	case *ForStmt:
		fmt.Fprintf(w, "%sForStmt\n", ind)
		if n.Init != nil {
			fmt.Fprintf(w, "%s  Init:\n", ind)
			fprintNode(w, n.Init, indent+2)
		}
		if n.Cond != nil {
			fmt.Fprintf(w, "%s  Cond:\n", ind)
			fprintNode(w, n.Cond, indent+2)
		}
		if n.Post != nil {
			fmt.Fprintf(w, "%s  Post:\n", ind)
			fprintNode(w, n.Post, indent+2)
		}
		fmt.Fprintf(w, "%s  Body:\n", ind)
		fprintNode(w, n.Body, indent+2)

	case *ForEachStmt:
		fmt.Fprintf(w, "%sForEachStmt var=%s\n", ind, n.VarName)
		fmt.Fprintf(w, "%s  ListExpr:\n", ind)
		fprintNode(w, n.ListExpr, indent+2)
		fmt.Fprintf(w, "%s  Body:\n", ind)
		fprintNode(w, n.Body, indent+2)

	case *ThrowStmt:
		fmt.Fprintf(w, "%sThrowStmt\n", ind)
		fprintNode(w, n.Expr, indent+1)

	case *BreakStmt:
		fmt.Fprintf(w, "%sBreakStmt\n", ind)

	case *TryStmt:
		fmt.Fprintf(w, "%sTryStmt\n", ind)
		fmt.Fprintf(w, "%s  Body:\n", ind)
		fprintNode(w, n.Body, indent+2)
		if n.CatchBody != nil {
			fmt.Fprintf(w, "%s  Catch var=%s:\n", ind, n.CatchName)
			fprintNode(w, n.CatchType, indent+2)
			fmt.Fprintf(w, "%s  CatchBody:\n", ind)
			fprintNode(w, n.CatchBody, indent+2)
		}

	case *IdentExpr:
		fmt.Fprintf(w, "%sIdent %s\n", ind, n.Name)

	case *IntLiteral:
		fmt.Fprintf(w, "%sIntLiteral %s\n", ind, n.Raw)

	case *FloatLiteral:
		fmt.Fprintf(w, "%sFloatLiteral %s\n", ind, n.Raw)

	case *StringLiteral:
		fmt.Fprintf(w, "%sStringLiteral %q\n", ind, n.Value)

	case *InterpolatedString:
		fmt.Fprintf(w, "%sInterpolatedString\n", ind)
		for _, part := range n.Parts {
			fprintNode(w, part, indent+1)
		}

	case *BytesLiteral:
		fmt.Fprintf(w, "%sBytesLiteral %d bytes\n", ind, len(n.Value))

	case *BoolLiteral:
		fmt.Fprintf(w, "%sBoolLiteral %v\n", ind, n.Value)

	case *StringTextPart:
		fmt.Fprintf(w, "%sStringTextPart %q\n", ind, n.Value)

	case *StringExprPart:
		fmt.Fprintf(w, "%sStringExprPart\n", ind)
		fprintNode(w, n.Expr, indent+1)

	case *NoneLiteral:
		fmt.Fprintf(w, "%sNoneLiteral\n", ind)

	case *SomeLiteral:
		fmt.Fprintf(w, "%sSomeLiteral\n", ind)
		fprintNode(w, n.Value, indent+1)

	case *ListLiteral:
		fmt.Fprintf(w, "%sListLiteral\n", ind)
		for _, el := range n.Elements {
			fprintNode(w, el, indent+1)
		}

	case *DictLiteral:
		fmt.Fprintf(w, "%sDictLiteral\n", ind)
		for _, entry := range n.Entries {
			fmt.Fprintf(w, "%s  Key %q:\n", ind, entry.Key)
			fprintNode(w, entry.Value, indent+2)
		}

	case *CallExpr:
		fmt.Fprintf(w, "%sCallExpr\n", ind)
		fmt.Fprintf(w, "%s  Callee:\n", ind)
		fprintNode(w, n.Callee, indent+2)
		if len(n.Args) > 0 {
			fmt.Fprintf(w, "%s  Args:\n", ind)
			for _, a := range n.Args {
				fprintNode(w, a, indent+2)
			}
		}

	case *IndexExpr:
		fmt.Fprintf(w, "%sIndexExpr\n", ind)
		fmt.Fprintf(w, "%s  X:\n", ind)
		fprintNode(w, n.X, indent+2)
		fmt.Fprintf(w, "%s  Index:\n", ind)
		fprintNode(w, n.Index, indent+2)

	case *MemberExpr:
		fmt.Fprintf(w, "%sMemberExpr name=%s\n", ind, n.Name)
		fmt.Fprintf(w, "%s  X:\n", ind)
		fprintNode(w, n.X, indent+2)

	case *BinaryExpr:
		fmt.Fprintf(w, "%sBinaryExpr op=%v\n", ind, n.Op)
		fmt.Fprintf(w, "%s  Left:\n", ind)
		fprintNode(w, n.Left, indent+2)
		fmt.Fprintf(w, "%s  Right:\n", ind)
		fprintNode(w, n.Right, indent+2)

	case *UnaryExpr:
		fmt.Fprintf(w, "%sUnaryExpr op=%v\n", ind, n.Op)
		fmt.Fprintf(w, "%s  X:\n", ind)
		fprintNode(w, n.X, indent+2)

	case *NamedArg:
		fmt.Fprintf(w, "%sNamedArg name=%s\n", ind, n.Name)
		fprintNode(w, n.Value, indent+1)

	case *FuncLiteral:
		fmt.Fprintf(w, "%sFuncLiteral\n", ind)
		if len(n.Params) > 0 {
			fmt.Fprintf(w, "%s  Params:\n", ind)
			for _, p := range n.Params {
				fprintNode(w, p, indent+2)
			}
		}
		fmt.Fprintf(w, "%s  Return:\n", ind)
		fprintNode(w, n.Return, indent+2)
		fmt.Fprintf(w, "%s  Body:\n", ind)
		fprintNode(w, n.Body, indent+2)

	default:
		fmt.Fprintf(w, "%s<unknown node %T>\n", ind, n)
	}
}
