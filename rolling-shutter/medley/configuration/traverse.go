package configuration

import (
	"encoding"
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
	sensitive     bool
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
	_, sensitive := tags["sensitive"]
	return tagOptions{
		setExplicitly: true,
		required:      required,
		sensitive:     sensitive,
	}
}

// GetSensitiveRecursive will return all sensitive config paths.
func GetSensitiveRecursive(root Config) map[string]any {
	sensitive := map[string]any{}
	execFn := func(n *node) error {
		if n.isBranch {
			return nil
		}
		opts := extractTags(n.fieldType)
		if opts.sensitive {
			sensitive[strings.ToLower(n.path)] = true
		}
		return nil
	}
	_ = traverseRecursive(root, execFn, true)
	return sensitive
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
			required[strings.ToLower(n.path)] = true
		}
		return nil
	}
	_ = traverseRecursive(root, execFn, true)
	return required
}

func SetDefaultValuesRecursive(root Config, excludePaths []string) error {
	exPth := map[string]bool{}
	for _, p := range excludePaths {
		exPth[p] = true
	}
	execFn := func(n *node) error {
		_, excluded := exPth[strings.ToLower(n.path)]
		if !n.isBranch || excluded {
			return nil
		}
		return n.config.SetDefaultValues()
	}
	// since parent nodes are executed after child nodes,
	// we need a second pass to reset excluded fields.
	// this is not efficient, but ok for config parsing
	execFnResetExcluded := func(n *node) error {
		_, excluded := exPth[strings.ToLower(n.path)]
		if !n.isBranch && excluded {
			if n.fieldValue.CanSet() {
				n.fieldValue.SetZero()
			}
			return nil
		}
		return nil
	}
	err := traverseRecursive(root, execFn, true)
	if err != nil {
		return err
	}
	return traverseRecursive(root, execFnResetExcluded, true)
}

func SetExampleValuesRecursive(root Config) error {
	execFn := func(n *node) error {
		if !n.isBranch {
			return nil
		}
		return n.config.SetExampleValues()
	}
	return traverseRecursive(root, execFn, true)
}

func GetEnvironmentVarsRecursive(root Config) map[string][]string {
	vars := map[string][]string{}
	execFn := func(n *node) error {
		if n.isBranch {
			return nil
		}
		vars[strings.ToLower(n.path)] = []string{
			// this is the old naming scheme, without the sub-config
			// qualifiers ("P2P", "ETHEREUM") and with a subcommand specific
			// prefix (e.g. KEYPER_CUSTOMBOOTSTRAPADDRESSES)
			strings.ToUpper(root.Name() + "_" + n.fieldType.Name),

			// full path with a generic prefix not dependend on the
			// subcommand executed
			// (e.g. SHUTTER_P2P_CUSTOMBOOTSTRAPADDRESSES)
			"SHUTTER_" + strings.ToUpper(strings.ReplaceAll(strings.ToLower(n.path), ".", "_")),
		}
		return nil
	}
	_ = traverseRecursive(root, execFn, true)
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
	err := traverseRecursive(root, execFn, true)
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

func traverseRecursive(root Config, exec execFunc, tailRecursive bool) error {
	rootNode := &node{
		isBranch:      true,
		previousNodes: []*node{},
		config:        root,
		path:          "",
		fieldType:     nil,
	}
	err := execRecursive(rootNode, exec, stopNever, tailRecursive)
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
func execRecursive(n *node, exec execFunc, stop stopFunc, tailRecursion bool) *execErr {
	if !tailRecursion {
		if err := exec(n); err != nil {
			return &execErr{
				err:  err,
				node: n,
			}
		}
	}
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
			nextPath += structField.Name
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
				if err := execRecursive(nextNode, exec, stop, tailRecursion); err != nil {
					return err
				}
			}
		}
	}
	if tailRecursion {
		if err := exec(n); err != nil {
			return &execErr{
				err:  err,
				node: n,
			}
		}
	}
	return nil
}

func UpsertDictForPath(d map[string]interface{}, path string) (map[string]interface{}, error) {
	var current map[string]interface{}
	if path == "" {
		return d, nil
	}
	for _, elem := range strings.Split(path, ".") {
		val, ok := d[elem]
		if !ok {
			current = map[string]interface{}{}
			d[elem] = current
			continue
		}
		current, ok = val.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("path dict (path=%s) has wrong type", path)
		}
	}
	return current, nil
}

func ToDict(root Config, redactPaths []string) (map[string]interface{}, error) {
	rdPth := map[string]bool{}
	for _, p := range redactPaths {
		rdPth[p] = true
	}
	out := map[string]interface{}{}
	execFn := func(n *node) error {
		if len(n.previousNodes) == 0 {
			return nil
		}
		if n.isBranch {
			_, err := UpsertDictForPath(out, n.path)
			if err != nil {
				return err
			}
		} else {
			subdict, err := UpsertDictForPath(out, n.previousNodes[len(n.previousNodes)-1].path)
			if err != nil {
				return err
			}

			_, redacted := rdPth[strings.ToLower(n.path)]
			var value any
			if redacted {
				value = "(sensitive)"
			} else {
				v := n.fieldValue.Interface()
				marshaller, ok := v.(encoding.TextMarshaler)
				if !ok {
					value = v
				} else {
					res, err := marshaller.MarshalText()
					if err != nil {
						return err
					}
					value = string(res)
				}
			}
			subdict[n.fieldType.Name] = value
		}
		return nil
	}
	err := traverseRecursive(root, execFn, false)
	return out, err
}
