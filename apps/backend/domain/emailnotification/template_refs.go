package emailnotification

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"text/template"
	"text/template/parse"
)

// errUnsupportedTemplateSyntax is returned for syntax whose key references
// cannot be read from the tree alone: range, with and variables change or
// alias the dot, and template/define move references into other trees
// (ADR-041).
var errUnsupportedTemplateSyntax = errors.New("unsupported email template syntax")

// checkReferencedKeys requires the keys referenced across the templates to
// be exactly keys (sorted), so neither an undeclared nor an unused key gets
// through.
func checkReferencedKeys(keys []string, templates ...*template.Template) error {
	referenced := map[string]struct{}{}
	for _, t := range templates {
		refs, err := referencedKeys(t)
		if err != nil {
			return err
		}
		maps.Copy(referenced, refs)
	}
	if got := slices.Sorted(maps.Keys(referenced)); !slices.Equal(got, keys) {
		return fmt.Errorf("%w: declared %q, referenced %q", errInvalidTemplateKeys, keys, got)
	}
	return nil
}

// referencedKeys returns the keys t references as {{.key}}, in every branch
// whether or not it runs for some params.
func referencedKeys(t *template.Template) (map[string]struct{}, error) {
	if len(t.Templates()) > 1 {
		return nil, fmt.Errorf("%w: define or block", errUnsupportedTemplateSyntax)
	}
	keys := map[string]struct{}{}
	if err := collectList(t.Root, keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func collectList(list *parse.ListNode, keys map[string]struct{}) error {
	if list == nil {
		return nil
	}
	for _, node := range list.Nodes {
		if err := collectNode(node, keys); err != nil {
			return err
		}
	}
	return nil
}

func collectNode(node parse.Node, keys map[string]struct{}) error {
	switch n := node.(type) {
	case *parse.TextNode, *parse.CommentNode:
		return nil
	case *parse.ActionNode:
		return collectPipe(n.Pipe, keys)
	case *parse.IfNode:
		if err := collectPipe(n.Pipe, keys); err != nil {
			return err
		}
		if err := collectList(n.List, keys); err != nil {
			return err
		}
		return collectList(n.ElseList, keys)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedTemplateSyntax, node)
	}
}

func collectPipe(pipe *parse.PipeNode, keys map[string]struct{}) error {
	if len(pipe.Decl) > 0 {
		return fmt.Errorf("%w: %s", errUnsupportedTemplateSyntax, pipe)
	}
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			if err := collectArg(arg, keys); err != nil {
				return err
			}
		}
	}
	return nil
}

func collectArg(arg parse.Node, keys map[string]struct{}) error {
	switch a := arg.(type) {
	case *parse.FieldNode:
		// Every value is a string, so .key.sub can never resolve.
		if len(a.Ident) != 1 {
			return fmt.Errorf("%w: %s", errUnsupportedTemplateSyntax, a)
		}
		keys[a.Ident[0]] = struct{}{}
		return nil
	case *parse.PipeNode:
		return collectPipe(a, keys)
	case *parse.IdentifierNode, *parse.StringNode, *parse.NumberNode, *parse.BoolNode:
		return nil
	default:
		return fmt.Errorf("%w: %s", errUnsupportedTemplateSyntax, arg)
	}
}
