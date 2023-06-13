package configuration

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func sliceToSet[T comparable](sl []T) map[T]bool {
	set := map[T]bool{}
	for _, e := range sl {
		set[e] = true
	}
	return set
}

type tagOptions struct {
	setExplicitly bool
	required      bool
}

func extractTags(field *reflect.StructField) tagOptions {
	if field == nil {
		return tagOptions{
			setExplicitly: false,
			required:      false,
		}
	}
	tagVal, ok := field.Tag.Lookup("shconfig")
	if !ok {
		return tagOptions{
			setExplicitly: false,
			required:      false,
		}
	}
	tags := sliceToSet(strings.Split(strings.TrimSpace(tagVal), ","))
	_, required := tags["required"]
	return tagOptions{
		setExplicitly: true,
		required:      required,
	}
}

// GetRequiredRecursive will return all required config paths
// A required field is a field that is not present as a default is considered required.
func GetRequiredRecursive(root Config) map[string]any {
	required := map[string]any{}
	execFn := func(n *node) error {
		if n.isBranch {
			return nil
		}
		opts := extractTags(n.fieldType)
		if opts.required {
			required[n.path] = true
		}
		return nil
	}
	_ = traverseRecursive(root, execFn)
	return required
}

func SetDefaultValuesRecursive(root Config, excludePaths []string) error {
	exPth := map[string]bool{}
	for _, p := range excludePaths {
		exPth[p] = true
	}
	execFn := func(n *node) error {
		_, excluded := exPth[n.path]
		if !n.isBranch || excluded {
			return nil
		}
		return n.config.SetDefaultValues()
	}
	// since parent nodes are executed after child nodes,
	// we need a second pass to reset excluded fields.
	// this is not efficient, but ok for config parsing
	execFnResetExcluded := func(n *node) error {
		_, excluded := exPth[n.path]
		if !n.isBranch && excluded {
			if n.fieldValue.CanSet() {
				n.fieldValue.SetZero()
			}
			return nil
		}
		return nil
	}
	err := traverseRecursive(root, execFn)
	if err != nil {
		return err
	}
	return traverseRecursive(root, execFnResetExcluded)
}

func SetExampleValuesRecursive(root Config) error {
	execFn := func(n *node) error {
		if !n.isBranch {
			return nil
		}
		return n.config.SetExampleValues()
	}
	return traverseRecursive(root, execFn)
}

func GetEnvironmentVarsRecursive(root Config) map[string][]string {
	vars := map[string][]string{}
	execFn := func(n *node) error {
		if n.isBranch {
			return nil
		}
		vars[n.path] = []string{
			// this is the old naming scheme, without the sub-config
			// qualifiers ("P2P", "ETHEREUM") and with a subcommand specific
			// prefix (e.g. KEYPER_CUSTOMBOOTSTRAPADDRESSES)
			strings.ToUpper(root.Name() + "_" + n.fieldType.Name),

			// full path with a generic prefix not dependend on the
			// subcommand executed
			// (e.g. SHUTTER_P2P_CUSTOMBOOTSTRAPADDRESSES)
			"SHUTTER_" + strings.ToUpper(strings.ReplaceAll(n.path, ".", "_")),
		}
		return nil
	}
	_ = traverseRecursive(root, execFn)
	return vars
}

func writeTOMLHeadersRecursive(root Config, w io.Writer) error {
	totalBytesWritten := 0
	execFn := func(n *node) error {
		if !n.isBranch {
			return nil
		}
		i, err := n.config.TOMLWriteHeader(w)
		totalBytesWritten += i
		return err
	}
	err := traverseRecursive(root, execFn)
	if err != nil {
		return err
	}
	if totalBytesWritten > 0 {
		_, err = fmt.Fprint(w, "\n\n")
	}
	return err
}

type (
	execFunc func(*node) error
	stopFunc func(*node) bool
	execErr  struct {
		err  error
		node *node
	}
	node struct {
		isBranch      bool
		previousNodes []*node
		path          string
		config        Config
		fieldType     *reflect.StructField
		fieldValue    *reflect.Value
	}
)

func (e *execErr) error() error {
	return errors.Wrapf(e.err, "error during recursion at path '%s'", e.node.path)
}

func traverseRecursive(root Config, exec execFunc) error {
	rootNode := &node{
		isBranch:      true,
		previousNodes: []*node{},
		config:        root,
		path:          "",
		fieldType:     nil,
	}
	err := execRecursive(rootNode, exec, stopNever)
	if err != nil {
		return err.error()
	}
	return nil
}

// stopAfterTopLevel executes only root and the root's child nodes.
func stopAfterTopLevel(n *node) bool {
	return len(n.previousNodes) > 2
}

// stopNever executes all nodes recursively including the root,
// leaves and branches.
func stopNever(_ *node) bool {
	return false
}

// execRecursive recursively traverses the config struct like a tree.
// The implementation is not optimized.
func execRecursive(n *node, exec execFunc, stop stopFunc) *execErr {
	// if the node is a branch, first handle potential subtrees.
	// this results in child nodes always being executed before their parent node
	if n.isBranch {
		newPath := append(n.previousNodes, n) //nolint: gocritic
		v := reflect.ValueOf(n.config)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		// Invalid kind means the nested config
		// struct is the null value (not initialized).
		// This happens when a parent config was not
		// initialized and the nested config has nil value.
		if v.Kind() == reflect.Invalid {
			return &execErr{
				err: errors.New(
					"config struct has not been initialized properly." +
						"nested config for path %s has nil value",
				),
				node: n,
			}
		}
		numFields := v.NumField()

		for i := 0; i < numFields; i++ {
			fld := v.Field(i)
			structField := v.Type().Field(i)
			if !structField.IsExported() {
				continue
			}

			nextPath := n.path
			if nextPath != "" {
				nextPath += "."
			}
			nextPath += strings.ToLower(structField.Name)
			nextNode := &node{
				isBranch:      false,
				previousNodes: newPath,
				path:          nextPath,
				config:        n.config,
				fieldType:     &structField,
				fieldValue:    &fld,
			}
			fieldVal := fld.Interface()
			nestedConfigVal, isNestedConfig := fieldVal.(Config)
			if isNestedConfig {
				nextNode.isBranch = true
				nextNode.config = nestedConfigVal
			}
			// if the stop function returns false,
			// we can dive into the next subtree
			if !stop(nextNode) {
				if err := execRecursive(nextNode, exec, stop); err != nil {
					return err
				}
			}
		}
	}
	if err := exec(n); err != nil {
		return &execErr{
			err:  err,
			node: n,
		}
	}
	return nil
}
