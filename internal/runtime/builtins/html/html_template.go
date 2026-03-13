package html

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

// --- Template AST node types ---

type nodeKind int

const (
	nodeText nodeKind = iota
	nodeExpr
	nodeIf
	nodeFor
	nodeBlock
	nodeExtends
	nodeInclude
)

type templateNode struct {
	kind        nodeKind
	text        string
	expr        string
	filter      string
	varName     string
	iterVar     string
	iterVar2    string
	children    []*templateNode
	elseBody    []*templateNode
	elifClauses []elifClause
}

type elifClause struct {
	expr     string
	children []*templateNode
}

// --- Compiled template ---

type compiledTemplate struct {
	name   string
	nodes  []*templateNode
	parent string
	blocks map[string][]*templateNode
	mtime  time.Time
}

// --- Engine ---

type engineHandle struct {
	mu      sync.RWMutex
	dir     string
	cache   map[string]*compiledTemplate
	devMode bool
}

func init() {
	registerNewEngine()
	registerEngineRender()
	registerEngineCompile()
	registerEngineSetDevMode()
	registerTemplateRender()
}

// --- Engine handle storage ---

var (
	engineStore   = make(map[uint64]*engineHandle)
	engineCounter uint64
)

func encodeEnginePtr(h *engineHandle) []byte {
	engineCounter++
	id := engineCounter
	engineStore[id] = h
	b := make([]byte, 8)
	b[0] = byte(id)
	b[1] = byte(id >> 8)
	b[2] = byte(id >> 16)
	b[3] = byte(id >> 24)
	b[4] = byte(id >> 32)
	b[5] = byte(id >> 40)
	b[6] = byte(id >> 48)
	b[7] = byte(id >> 56)
	return b
}

func decodeEnginePtr(b []byte) *engineHandle {
	if len(b) != 8 {
		return nil
	}
	id := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	return engineStore[id]
}

// --- Template handle storage ---

var (
	tplStore   = make(map[uint64]*compiledTemplate)
	tplCounter uint64
)

func encodeTplPtr(t *compiledTemplate) []byte {
	tplCounter++
	id := tplCounter
	tplStore[id] = t
	b := make([]byte, 8)
	b[0] = byte(id)
	b[1] = byte(id >> 8)
	b[2] = byte(id >> 16)
	b[3] = byte(id >> 24)
	b[4] = byte(id >> 32)
	b[5] = byte(id >> 40)
	b[6] = byte(id >> 48)
	b[7] = byte(id >> 56)
	return b
}

func decodeTplPtr(b []byte) *compiledTemplate {
	if len(b) != 8 {
		return nil
	}
	id := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	return tplStore[id]
}

func registerNewEngine() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLNewEngine,
			Name:       "__builtin_html_new_engine",
			Arity:      2,
			ParamNames: []string{"dir", "opts"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeDict},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dirVal := args[0].(value.Value)
			optsVal := args[1].(value.Value)

			dir := dirVal.Str
			if !filepath.IsAbs(dir) {
				root := env.ExecRoot()
				if root == "" {
					root, _ = os.Getwd()
				}
				if root != "" {
					dir = filepath.Join(root, dir)
				}
			}

			devMode := false
			if optsVal.Kind == value.KindDict {
				if dm, ok := optsVal.Dict["devMode"]; ok && dm.Kind == value.KindBool {
					devMode = dm.Bool
				}
			}

			eng := &engineHandle{
				dir:     dir,
				cache:   make(map[string]*compiledTemplate),
				devMode: devMode,
			}

			return value.Value{Kind: value.KindBytes, Bytes: encodeEnginePtr(eng)}, nil
		},
	})
}

func registerEngineRender() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLEngineRender,
			Name:       "__builtin_html_engine_render",
			Arity:      3,
			ParamNames: []string{"handle", "name", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeDict},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			nameVal := args[1].(value.Value)
			dataVal := args[2].(value.Value)

			eng := decodeEnginePtr(handleVal.Bytes)
			if eng == nil {
				return value.Value{}, fmt.Errorf("html.engine.render: invalid engine handle")
			}

			tpl, err := eng.getTemplate(nameVal.Str)
			if err != nil {
				return value.Value{}, err
			}

			result, err := renderTemplate(eng, tpl, dataVal)
			if err != nil {
				return value.Value{}, err
			}

			return value.Str(result), nil
		},
	})
}

func registerEngineCompile() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLEngineCompile,
			Name:       "__builtin_html_engine_compile",
			Arity:      2,
			ParamNames: []string{"handle", "name"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			nameVal := args[1].(value.Value)

			eng := decodeEnginePtr(handleVal.Bytes)
			if eng == nil {
				return value.Value{}, fmt.Errorf("html.engine.compile: invalid engine handle")
			}

			tpl, err := eng.getTemplate(nameVal.Str)
			if err != nil {
				return value.Value{}, err
			}

			return value.Value{Kind: value.KindBytes, Bytes: encodeTplPtr(tpl)}, nil
		},
	})
}

func registerEngineSetDevMode() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLEngineSetDevMode,
			Name:       "__builtin_html_engine_set_dev_mode",
			Arity:      2,
			ParamNames: []string{"handle", "enabled"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBool},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			enabledVal := args[1].(value.Value)

			eng := decodeEnginePtr(handleVal.Bytes)
			if eng == nil {
				return value.Value{}, fmt.Errorf("html.engine.set_dev_mode: invalid engine handle")
			}

			eng.mu.Lock()
			eng.devMode = enabledVal.Bool
			eng.mu.Unlock()

			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerTemplateRender() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLTemplateRender,
			Name:       "__builtin_html_template_render",
			Arity:      2,
			ParamNames: []string{"handle", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeDict},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			dataVal := args[1].(value.Value)

			tpl := decodeTplPtr(handleVal.Bytes)
			if tpl == nil {
				return value.Value{}, fmt.Errorf("html.template.render: invalid template handle")
			}

			result, err := renderTemplate(nil, tpl, dataVal)
			if err != nil {
				return value.Value{}, err
			}

			return value.Str(result), nil
		},
	})
}

// --- Engine methods ---

func (eng *engineHandle) getTemplate(name string) (*compiledTemplate, error) {
	eng.mu.RLock()
	cached, ok := eng.cache[name]
	devMode := eng.devMode
	eng.mu.RUnlock()

	if ok && !devMode {
		return cached, nil
	}

	path := filepath.Join(eng.dir, name)

	if ok && devMode {
		info, err := os.Stat(path)
		if err == nil && !info.ModTime().After(cached.mtime) {
			return cached, nil
		}
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", name, err)
	}

	info, _ := os.Stat(path)
	var mtime time.Time
	if info != nil {
		mtime = info.ModTime()
	}

	tpl, err := parseTemplate(name, string(src))
	if err != nil {
		return nil, err
	}
	tpl.mtime = mtime

	eng.mu.Lock()
	eng.cache[name] = tpl
	eng.mu.Unlock()

	return tpl, nil
}

// --- Template parser ---

func parseTemplate(name, src string) (*compiledTemplate, error) {
	nodes, err := parseNodes(name, src, 0)
	if err != nil {
		return nil, err
	}

	tpl := &compiledTemplate{
		name:   name,
		nodes:  nodes,
		blocks: make(map[string][]*templateNode),
	}

	for _, n := range nodes {
		if n.kind == nodeExtends {
			tpl.parent = n.text
		}
		if n.kind == nodeBlock {
			tpl.blocks[n.text] = n.children
		}
	}

	return tpl, nil
}

func parseNodes(name, src string, depth int) ([]*templateNode, error) {
	var nodes []*templateNode
	i := 0

	for i < len(src) {
		// Look for {{ or {% or {#
		tagStart := strings.Index(src[i:], "{")
		if tagStart == -1 {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:]})
			break
		}
		tagStart += i

		if tagStart+1 >= len(src) {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:]})
			break
		}

		nextChar := src[tagStart+1]

		if nextChar == '{' {
			// Expression: {{ ... }}
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "}}")
			if end == -1 {
				return nil, fmt.Errorf("template %q: unclosed {{ at position %d", name, tagStart)
			}
			end += tagStart + 2
			raw := strings.TrimSpace(src[tagStart+2 : end])
			expr, filter := parseExprFilter(raw)
			nodes = append(nodes, &templateNode{kind: nodeExpr, expr: expr, filter: filter})
			i = end + 2
		} else if nextChar == '%' {
			// Tag: {% ... %}
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "%}")
			if end == -1 {
				return nil, fmt.Errorf("template %q: unclosed {%% at position %d", name, tagStart)
			}
			end += tagStart + 2
			tag := strings.TrimSpace(src[tagStart+2 : end])
			i = end + 2

			node, rest, err := parseTag(name, tag, src[i:], depth)
			if err != nil {
				return nil, err
			}
			if node != nil {
				nodes = append(nodes, node)
			}
			i = len(src) - len(rest)
		} else if nextChar == '#' {
			// Comment: {# ... #}
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "#}")
			if end == -1 {
				return nil, fmt.Errorf("template %q: unclosed {# at position %d", name, tagStart)
			}
			i = tagStart + 2 + end + 2
		} else {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i : tagStart+1]})
			i = tagStart + 1
		}
	}

	return nodes, nil
}

func parseExprFilter(raw string) (string, string) {
	parts := strings.SplitN(raw, "|", 2)
	expr := strings.TrimSpace(parts[0])
	filter := ""
	if len(parts) > 1 {
		filter = strings.TrimSpace(parts[1])
	}
	return expr, filter
}

func parseTag(name, tag, rest string, depth int) (*templateNode, string, error) {
	parts := splitTagParts(tag)
	if len(parts) == 0 {
		return nil, rest, nil
	}

	switch parts[0] {
	case "if":
		return parseIfTag(name, parts, rest, depth)
	case "for":
		return parseForTag(name, parts, rest, depth)
	case "block":
		return parseBlockTag(name, parts, rest, depth)
	case "extends":
		if len(parts) < 2 {
			return nil, rest, fmt.Errorf("template %q: extends requires a template name", name)
		}
		parentName := unquote(parts[1])
		return &templateNode{kind: nodeExtends, text: parentName}, rest, nil
	case "include":
		if len(parts) < 2 {
			return nil, rest, fmt.Errorf("template %q: include requires a template name", name)
		}
		includeName := unquote(parts[1])
		node := &templateNode{kind: nodeInclude, text: includeName}
		// Parse optional "with key=val" pairs
		for i := 2; i < len(parts); i++ {
			if parts[i] == "with" {
				continue
			}
			node.varName += parts[i] + " "
		}
		node.varName = strings.TrimSpace(node.varName)
		return node, rest, nil
	case "endif", "endfor", "endblock", "else", "elif":
		return nil, tag + "%}" + rest, nil
	default:
		return nil, rest, fmt.Errorf("template %q: unknown tag %q", name, parts[0])
	}
}

func parseIfTag(name string, parts []string, rest string, depth int) (*templateNode, string, error) {
	expr := strings.Join(parts[1:], " ")
	node := &templateNode{kind: nodeIf, expr: expr}

	body, remaining, endTag, err := collectUntilTags(name, rest, depth, "endif", "else", "elif")
	if err != nil {
		return nil, "", err
	}
	node.children = body

	for {
		if endTag == "endif" {
			break
		}
		if endTag == "else" {
			elseBody, remaining2, _, err := collectUntilTags(name, remaining, depth, "endif")
			if err != nil {
				return nil, "", err
			}
			node.elseBody = elseBody
			remaining = remaining2
			break
		}
		if endTag == "elif" {
			elifExpr, elifRemaining := extractElifExpr(remaining)
			elifBody, remaining2, nextTag, err := collectUntilTags(name, elifRemaining, depth, "endif", "else", "elif")
			if err != nil {
				return nil, "", err
			}
			node.elifClauses = append(node.elifClauses, elifClause{expr: elifExpr, children: elifBody})
			remaining = remaining2
			endTag = nextTag
		}
	}

	return node, remaining, nil
}

func extractElifExpr(rest string) (string, string) {
	// The elif expression was already consumed by the tag parser, so we need to
	// handle this differently. The "rest" here starts right after "%}" of the elif tag.
	// We actually need to look at what was passed — the tag content is in the endTag mechanism.
	// Let me reconsider: collectUntilTags returns the endTag keyword and remaining starts after %}.
	// But we need the full tag content for elif. Let me adjust.
	return rest, ""
}

func parseForTag(name string, parts []string, rest string, depth int) (*templateNode, string, error) {
	// {% for item in list %} or {% for key, value in dict %}
	node := &templateNode{kind: nodeFor}

	inIdx := -1
	for i, p := range parts {
		if p == "in" {
			inIdx = i
			break
		}
	}
	if inIdx == -1 || inIdx < 2 {
		return nil, rest, fmt.Errorf("template %q: invalid for syntax", name)
	}

	vars := strings.Join(parts[1:inIdx], " ")
	varParts := strings.Split(vars, ",")
	node.iterVar = strings.TrimSpace(varParts[0])
	if len(varParts) > 1 {
		node.iterVar2 = strings.TrimSpace(varParts[1])
	}
	node.expr = strings.Join(parts[inIdx+1:], " ")

	body, remaining, _, err := collectUntilTags(name, rest, depth, "endfor")
	if err != nil {
		return nil, "", err
	}
	node.children = body
	return node, remaining, nil
}

func parseBlockTag(name string, parts []string, rest string, depth int) (*templateNode, string, error) {
	if len(parts) < 2 {
		return nil, rest, fmt.Errorf("template %q: block requires a name", name)
	}
	node := &templateNode{kind: nodeBlock, text: parts[1]}

	body, remaining, _, err := collectUntilTags(name, rest, depth, "endblock")
	if err != nil {
		return nil, "", err
	}
	node.children = body
	return node, remaining, nil
}

func collectUntilTags(name, src string, depth int, endTags ...string) ([]*templateNode, string, string, error) {
	var nodes []*templateNode
	i := 0

	for i < len(src) {
		tagStart := strings.Index(src[i:], "{")
		if tagStart == -1 {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:]})
			return nil, "", "", fmt.Errorf("template %q: unclosed block (expected one of: %v)", name, endTags)
		}
		tagStart += i

		if tagStart+1 >= len(src) {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:]})
			return nil, "", "", fmt.Errorf("template %q: unclosed block", name)
		}

		nextChar := src[tagStart+1]

		if nextChar == '{' {
			// Expression
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "}}")
			if end == -1 {
				return nil, "", "", fmt.Errorf("template %q: unclosed {{", name)
			}
			end += tagStart + 2
			raw := strings.TrimSpace(src[tagStart+2 : end])
			expr, filter := parseExprFilter(raw)
			nodes = append(nodes, &templateNode{kind: nodeExpr, expr: expr, filter: filter})
			i = end + 2
		} else if nextChar == '%' {
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "%}")
			if end == -1 {
				return nil, "", "", fmt.Errorf("template %q: unclosed {%%", name)
			}
			end += tagStart + 2
			tag := strings.TrimSpace(src[tagStart+2 : end])
			tagParts := splitTagParts(tag)
			i = end + 2

			if len(tagParts) > 0 {
				for _, et := range endTags {
					if tagParts[0] == et {
						return nodes, src[i:], et, nil
					}
					if et == "elif" && tagParts[0] == "elif" {
						// For elif, we need to pass the expression along
						// Store it in the remaining so the caller can extract it
						elifExpr := strings.Join(tagParts[1:], " ")
						return nodes, elifExpr + "\x00" + src[i:], "elif", nil
					}
				}

				// Nested tag
				node, rest, err := parseTag(name, tag, src[i:], depth+1)
				if err != nil {
					return nil, "", "", err
				}
				if node != nil {
					nodes = append(nodes, node)
				}
				i = len(src) - len(rest)
			}
		} else if nextChar == '#' {
			if tagStart > i {
				nodes = append(nodes, &templateNode{kind: nodeText, text: src[i:tagStart]})
			}
			end := strings.Index(src[tagStart+2:], "#}")
			if end == -1 {
				return nil, "", "", fmt.Errorf("template %q: unclosed {#", name)
			}
			i = tagStart + 2 + end + 2
		} else {
			nodes = append(nodes, &templateNode{kind: nodeText, text: src[i : tagStart+1]})
			i = tagStart + 1
		}
	}

	return nil, "", "", fmt.Errorf("template %q: unclosed block (expected one of: %v)", name, endTags)
}

// --- Template renderer ---

func renderTemplate(eng *engineHandle, tpl *compiledTemplate, data value.Value) (string, error) {
	// Handle template inheritance
	if tpl.parent != "" && eng != nil {
		parentTpl, err := eng.getTemplate(tpl.parent)
		if err != nil {
			return "", fmt.Errorf("template %q: error loading parent %q: %w", tpl.name, tpl.parent, err)
		}
		mergedBlocks := make(map[string][]*templateNode)
		for k, v := range parentTpl.blocks {
			mergedBlocks[k] = v
		}
		for k, v := range tpl.blocks {
			mergedBlocks[k] = v
		}

		var buf strings.Builder
		ctx := &renderContext{
			data:   data,
			blocks: mergedBlocks,
			eng:    eng,
			scope:  make(map[string]value.Value),
		}
		if err := renderNodes(&buf, parentTpl.nodes, ctx); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	var buf strings.Builder
	ctx := &renderContext{
		data:   data,
		blocks: tpl.blocks,
		eng:    eng,
		scope:  make(map[string]value.Value),
	}
	if err := renderNodes(&buf, tpl.nodes, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type renderContext struct {
	data   value.Value
	blocks map[string][]*templateNode
	eng    *engineHandle
	scope  map[string]value.Value
}

func (ctx *renderContext) clone() *renderContext {
	newScope := make(map[string]value.Value, len(ctx.scope))
	for k, v := range ctx.scope {
		newScope[k] = v
	}
	return &renderContext{
		data:   ctx.data,
		blocks: ctx.blocks,
		eng:    ctx.eng,
		scope:  newScope,
	}
}

func renderNodes(buf *strings.Builder, nodes []*templateNode, ctx *renderContext) error {
	for _, n := range nodes {
		switch n.kind {
		case nodeText:
			buf.WriteString(n.text)
		case nodeExpr:
			val := evalExpr(n.expr, ctx)
			text := valueToString(val)
			if n.filter == "raw" {
				buf.WriteString(text)
			} else {
				text = applyFilter(text, n.filter)
				buf.WriteString(escapeHTML(text))
			}
		case nodeIf:
			val := evalExpr(n.expr, ctx)
			if isTruthy(val) {
				if err := renderNodes(buf, n.children, ctx); err != nil {
					return err
				}
			} else {
				matched := false
				for _, elif := range n.elifClauses {
					elifVal := evalExpr(elif.expr, ctx)
					if isTruthy(elifVal) {
						if err := renderNodes(buf, elif.children, ctx); err != nil {
							return err
						}
						matched = true
						break
					}
				}
				if !matched && n.elseBody != nil {
					if err := renderNodes(buf, n.elseBody, ctx); err != nil {
						return err
					}
				}
			}
		case nodeFor:
			collection := evalExpr(n.expr, ctx)
			childCtx := ctx.clone()
			if collection.Kind == value.KindList {
				for _, item := range collection.List {
					childCtx.scope[n.iterVar] = item
					if err := renderNodes(buf, n.children, childCtx); err != nil {
						return err
					}
				}
			} else if collection.Kind == value.KindDict {
				for k, v := range collection.Dict {
					if n.iterVar2 != "" {
						childCtx.scope[n.iterVar] = value.Str(k)
						childCtx.scope[n.iterVar2] = v
					} else {
						childCtx.scope[n.iterVar] = value.Str(k)
					}
					if err := renderNodes(buf, n.children, childCtx); err != nil {
						return err
					}
				}
			}
		case nodeBlock:
			blockNodes := n.children
			if override, ok := ctx.blocks[n.text]; ok {
				blockNodes = override
			}
			if err := renderNodes(buf, blockNodes, ctx); err != nil {
				return err
			}
		case nodeExtends:
			// handled at top level
		case nodeInclude:
			if ctx.eng != nil {
				incTpl, err := ctx.eng.getTemplate(n.text)
				if err != nil {
					return fmt.Errorf("include %q: %w", n.text, err)
				}
				incCtx := ctx.clone()
				// Parse "with" vars if present
				if n.varName != "" {
					parseWithVars(n.varName, incCtx, ctx)
				}
				if err := renderNodes(buf, incTpl.nodes, incCtx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseWithVars(withStr string, dest, src *renderContext) {
	pairs := strings.Fields(withStr)
	for _, pair := range pairs {
		eqIdx := strings.Index(pair, "=")
		if eqIdx > 0 {
			key := pair[:eqIdx]
			valExpr := pair[eqIdx+1:]
			dest.scope[key] = evalExpr(valExpr, src)
		}
	}
}

// --- Expression evaluator ---

func evalExpr(expr string, ctx *renderContext) value.Value {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return value.Str("")
	}

	// String literal
	if (strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`)) ||
		(strings.HasPrefix(expr, `'`) && strings.HasSuffix(expr, `'`)) {
		return value.Str(expr[1 : len(expr)-1])
	}

	// Integer literal
	if isIntLiteral(expr) {
		var n int64
		fmt.Sscanf(expr, "%d", &n)
		return value.Int(n)
	}

	// Boolean literals
	if expr == "true" {
		return value.Bool(true)
	}
	if expr == "false" {
		return value.Bool(false)
	}
	if expr == "none" {
		return value.None()
	}

	// Comparison operators
	for _, op := range []string{">=", "<=", "!=", "==", ">", "<"} {
		if idx := findOperator(expr, op); idx >= 0 {
			left := evalExpr(expr[:idx], ctx)
			right := evalExpr(expr[idx+len(op):], ctx)
			return evalComparison(left, right, op)
		}
	}

	// Logical operators
	if idx := findKeywordOp(expr, "and"); idx >= 0 {
		left := evalExpr(expr[:idx], ctx)
		if !isTruthy(left) {
			return value.Bool(false)
		}
		right := evalExpr(expr[idx+3:], ctx)
		return value.Bool(isTruthy(right))
	}
	if idx := findKeywordOp(expr, "or"); idx >= 0 {
		left := evalExpr(expr[:idx], ctx)
		if isTruthy(left) {
			return value.Bool(true)
		}
		right := evalExpr(expr[idx+2:], ctx)
		return value.Bool(isTruthy(right))
	}

	// Not operator
	if strings.HasPrefix(expr, "not ") {
		inner := evalExpr(expr[4:], ctx)
		return value.Bool(!isTruthy(inner))
	}

	// Arithmetic: + operator (also string concatenation)
	if idx := findOperator(expr, "+"); idx >= 0 {
		left := evalExpr(expr[:idx], ctx)
		right := evalExpr(expr[idx+1:], ctx)
		if left.Kind == value.KindString || right.Kind == value.KindString {
			return value.Str(valueToString(left) + valueToString(right))
		}
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Int(left.Int + right.Int)
		}
	}

	// Arithmetic: - operator
	if idx := findOperatorReverse(expr, "-"); idx > 0 {
		left := evalExpr(expr[:idx], ctx)
		right := evalExpr(expr[idx+1:], ctx)
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Int(left.Int - right.Int)
		}
	}

	// Arithmetic: * operator
	if idx := findOperator(expr, "*"); idx >= 0 {
		left := evalExpr(expr[:idx], ctx)
		right := evalExpr(expr[idx+1:], ctx)
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Int(left.Int * right.Int)
		}
	}

	// Dict access: expr["key"]
	if bracketIdx := strings.Index(expr, "["); bracketIdx > 0 {
		if strings.HasSuffix(expr, "]") {
			base := evalExpr(expr[:bracketIdx], ctx)
			keyStr := strings.TrimSpace(expr[bracketIdx+1 : len(expr)-1])
			keyVal := evalExpr(keyStr, ctx)
			if base.Kind == value.KindDict && keyVal.Kind == value.KindString {
				if v, ok := base.Dict[keyVal.Str]; ok {
					return v
				}
			}
			return value.Str("")
		}
	}

	// Dot access: var.field.field
	return lookupDotPath(expr, ctx)
}

func lookupDotPath(path string, ctx *renderContext) value.Value {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return value.Str("")
	}

	// First resolve the root variable
	root := strings.TrimSpace(parts[0])
	var current value.Value

	// Check loop scope first
	if v, ok := ctx.scope[root]; ok {
		current = v
	} else if ctx.data.Kind == value.KindDict {
		if v, ok := ctx.data.Dict[root]; ok {
			current = v
		} else {
			return value.Str("")
		}
	} else {
		return value.Str("")
	}

	// Resolve remaining parts
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		switch current.Kind {
		case value.KindDict:
			if v, ok := current.Dict[part]; ok {
				current = v
			} else {
				return value.Str("")
			}
		case value.KindStruct:
			if current.Struct == nil {
				return value.Str("")
			}
			found := false
			for i, f := range current.Struct.Fields {
				_ = i
				// Struct field access by index requires type info which we don't have here.
				// Fall back to treating struct as dict-like if fields are named.
				_ = f
			}
			if !found {
				return value.Str("")
			}
		default:
			return value.Str("")
		}
	}

	return current
}

func evalComparison(left, right value.Value, op string) value.Value {
	switch op {
	case "==":
		return value.Bool(valuesEqual(left, right))
	case "!=":
		return value.Bool(!valuesEqual(left, right))
	case ">":
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Bool(left.Int > right.Int)
		}
		return value.Bool(false)
	case "<":
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Bool(left.Int < right.Int)
		}
		return value.Bool(false)
	case ">=":
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Bool(left.Int >= right.Int)
		}
		return value.Bool(false)
	case "<=":
		if left.Kind == value.KindInt && right.Kind == value.KindInt {
			return value.Bool(left.Int <= right.Int)
		}
		return value.Bool(false)
	}
	return value.Bool(false)
}

func valuesEqual(a, b value.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case value.KindInt:
		return a.Int == b.Int
	case value.KindFloat:
		return a.Float == b.Float
	case value.KindString:
		return a.Str == b.Str
	case value.KindBool:
		return a.Bool == b.Bool
	default:
		return false
	}
}

func isTruthy(v value.Value) bool {
	switch v.Kind {
	case value.KindBool:
		return v.Bool
	case value.KindInt:
		return v.Int != 0
	case value.KindString:
		return v.Str != ""
	case value.KindList:
		return len(v.List) > 0
	case value.KindDict:
		return len(v.Dict) > 0
	case value.KindOptional:
		return v.Optional != nil && v.Optional.IsSome
	default:
		return true
	}
}

func valueToString(v value.Value) string {
	switch v.Kind {
	case value.KindString:
		return v.Str
	case value.KindInt:
		return fmt.Sprintf("%d", v.Int)
	case value.KindFloat:
		return fmt.Sprintf("%g", v.Float)
	case value.KindBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case value.KindOptional:
		if v.Optional == nil || !v.Optional.IsSome {
			return ""
		}
		return valueToString(v.Optional.Value)
	default:
		return v.String()
	}
}

func applyFilter(text, filter string) string {
	switch filter {
	case "upper":
		return strings.ToUpper(text)
	case "lower":
		return strings.ToLower(text)
	case "trim":
		return strings.TrimSpace(text)
	case "":
		return text
	default:
		// Check for default(val) filter
		if strings.HasPrefix(filter, "default(") && strings.HasSuffix(filter, ")") {
			if text == "" {
				def := filter[8 : len(filter)-1]
				return unquote(def)
			}
		}
		return text
	}
}

// --- Helpers ---

func splitTagParts(tag string) []string {
	return strings.Fields(tag)
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func isIntLiteral(s string) bool {
	if len(s) == 0 {
		return false
	}
	start := 0
	if s[0] == '-' {
		if len(s) == 1 {
			return false
		}
		start = 1
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func findOperator(expr, op string) int {
	depth := 0
	for i := 0; i < len(expr)-len(op)+1; i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '"', '\'':
			q := expr[i]
			i++
			for i < len(expr) && expr[i] != q {
				i++
			}
		default:
			if depth == 0 && expr[i:i+len(op)] == op {
				if len(op) == 1 && (op == ">" || op == "<") {
					if i+1 < len(expr) && expr[i+1] == '=' {
						continue
					}
					if i > 0 && (expr[i-1] == '!' || expr[i-1] == '<' || expr[i-1] == '>') {
						continue
					}
				}
				return i
			}
		}
	}
	return -1
}

func findOperatorReverse(expr, op string) int {
	depth := 0
	for i := len(expr) - 1; i >= len(op)-1; i-- {
		switch expr[i] {
		case ')':
			depth++
		case '(':
			depth--
		default:
			if depth == 0 && i-len(op)+1 >= 0 && expr[i-len(op)+1:i+1] == op {
				return i - len(op) + 1
			}
		}
	}
	return -1
}

func findKeywordOp(expr, keyword string) int {
	kLen := len(keyword)
	for i := 1; i < len(expr)-kLen; i++ {
		if expr[i:i+kLen] == keyword {
			if expr[i-1] == ' ' && i+kLen < len(expr) && expr[i+kLen] == ' ' {
				return i
			}
		}
	}
	return -1
}
