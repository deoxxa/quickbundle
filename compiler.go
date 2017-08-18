package main

import (
	"context"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync/atomic"
	"text/template"
)

type Compiler struct {
	lastID  uint64
	modules map[string]*Module
}

func NewCompiler() *Compiler {
	return &Compiler{
		modules: make(map[string]*Module),
	}
}

func (c *Compiler) Load(ctx context.Context, path string) (*Module, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	path = p

	if _, ok := c.modules[path]; !ok {
		input, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		h := sha1.New()
		if _, err := h.Write(input); err != nil {
			return nil, err
		}

		m := newModule(atomic.AddUint64(&c.lastID, 1), path, h.Sum(make([]byte, sha1.Size)), input)

		c.modules[path] = m

		for _, dep := range m.Dependencies() {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(path), dep))
			if err != nil {
				return nil, err
			}

			m.deps[dep] = abs
			m.uses[abs] = true

			other, err := c.Load(ctx, abs)
			if err != nil {
				return nil, err
			}

			other.refs[path] = true
		}
	}

	return c.modules[path], nil
}

//
// heavily inspired by https://github.com/substack/browser-pack/blob/master/prelude.js
//
var tpl = template.Must(template.New("module").Parse(`
{{- $Root := . -}}
(function outer(modules, cache, entry) {
  function __require(name) {
    if (!cache[name]) {
      var m = cache[name] = { exports: {} };

      modules[name][0].call(m.exports, function(requested) {
        var id = modules[name][1][requested];
        return __require(id ? id : requested);
      }, m, m.exports, outer, modules, cache, entry);
    }

    return cache[name].exports;
  }

  for (var i = 0; i < entry.length; i++) {
    __require(entry[i]);
  }

  return __require;
})({
{{- range $Module := .Modules}}
  "{{$Module.ID}}": [
    function(require, module, exports) {
{{$Module.Output}}
    },
    {
  {{- range $Name, $Path := $Module.Deps}}
      "{{$Name}}": {{(index $Root.Modules $Path).ID}},
  {{- end}}
    },
  ],
{{- end}}
}, {}, [ {{.Entry}} ]);`))

func (c *Compiler) BundleModule(ctx context.Context, m *Module, wr io.Writer) error {
	return tpl.Execute(wr, map[string]interface{}{
		"Modules": c.modules,
		"Entry":   m.ID(),
	})
}
