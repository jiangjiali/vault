package hclutil

import (
	"fmt"

	"github.com/jiangjiali/vault/sdk/helper/hclutil/hcl/hcl/ast"
	"github.com/jiangjiali/vault/sdk/helper/multierror"
)

// CheckHCLKeys checks whether the keys in the AST list contains any of the valid keys provided.
func CheckHCLKeys(node ast.Node, valid []string) error {
	var list *ast.ObjectList
	switch n := node.(type) {
	case *ast.ObjectList:
		list = n
	case *ast.ObjectType:
		list = n.List
	default:
		return fmt.Errorf("cannot check HCL keys of type %T", n)
	}

	validMap := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		validMap[v] = struct{}{}
	}

	var result error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			result = multierror.Append(result, fmt.Errorf("invalid key %q on line %d", key, item.Assign.Line))
		}
	}

	return result
}
