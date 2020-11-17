/*
Copyright 2020 Devtron Labs Pvt Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package language

import (
	json2 "encoding/json"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	parser2 "github.com/devtron-labs/inception/pkg/language/parser"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	yaml "sigs.k8s.io/yaml"
	"strings"
	"testing"
)

func TestKlangListener_handleNestedIf(t *testing.T) {
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nested if",
			fields: fields{
				input: `
name = "name";
if name == name && name == name && name == name && name == name || name == name { a=1+2; }
if 2==2 {
if 1==1 {
a = 2; 
}
else {
a=4;
}
} 
else { a = 6;}
b = 2 - 3;
c = 6+8;
d = 2 / 3.3;
e = 3.3 * 2.2;
f = "abc" + name;`,
				values: map[string]valHolder{
					"a": {
						dataType: INT,
						name:     "a",
						value:    int64(2),
					},
					"b": {
						dataType: INT,
						name:     "b",
						value:    int64(-1),
					},
					"c": {
						dataType: INT,
						name:     "c",
						value:    int64(14),
					},
					"d": {
						dataType: FLOAT,
						name:     "d",
						value:    0.6060606060606061,
					},
					"e": {
						dataType: FLOAT,
						name:     "e",
						value:    7.26,
					},
					"f": {
						dataType: STRING,
						name:     "f",
						value:    "abcname",
					},
					"name": {
						dataType: STRING,
						name:     "name",
						value:    "name",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if diff := compare(r.Values(), tt.fields.values); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleWhile(t *testing.T) {
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "While",
			fields: fields{
				input: `
x = 0;
while x < 2 {
x = x+1;
}`,
				values: map[string]valHolder{
					"x": {
						dataType: "INT",
						name:     "x",
						value:    int64(2),
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if diff := compare(tt.fields.values, r.values); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleJsonSelect(t *testing.T) {
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Json Select",
			fields: fields{
				input: `
x = {"name":{"first":"abc","last":"def"}};
y = jsonSelect(x, "name.last");
`,
				values: map[string]valHolder{
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    "def",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["y"]; ok {
				m := map[string]valHolder{
					"y": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleJsonEdit(t *testing.T) {
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Json Edit",
			fields: fields{
				input: `
x = {"name":{"first":"abc","last":"def"}};
jsonEdit(x, "name.first", "xyz");
`,
				values: map[string]valHolder{
					"x": {
						dataType: "STRING",
						name:     "x",
						value:    "{\"name\":{\"first\":\"xyz\",\"last\":\"def\"}}",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["x"]; ok {
				m := map[string]valHolder{
					"x": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlSelect(t *testing.T) {
	y := `
name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
y = yamlSelect(x, "name.last");
`
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Yaml Select",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    "def",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["y"]; ok {
				m := map[string]valHolder{
					"y": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlMultiSelect(t *testing.T) {
	y := `
name:
  first: ghi
  last: jkl
---
name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
y = yamlSelect(x, "name.last", 1);
`
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Yaml Multi Select",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    "def",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["y"]; ok {
				m := map[string]valHolder{
					"y": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlEdit(t *testing.T) {
	y := `
name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
yamlEdit(x, "name.first", "xyz");
`
	o := `name:
  first: xyz
  last: def
`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Yaml Edit",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"x": {
						dataType: "STRING",
						name:     "x",
						value:    o,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["x"]; ok {
				m := map[string]valHolder{
					"x": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlMultiEdit(t *testing.T) {
	y := `
name:
  first: efg
  last: hij
---
name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
yamlEdit(x, "name.first", "xyz", 1);
`
	o := `
name:
  first: efg
  last: hij
---
name:
  first: xyz
  last: def
`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Yaml Edit",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"x": {
						dataType: "STRING",
						name:     "x",
						value:    o,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["x"]; ok {
				m := map[string]valHolder{
					"x": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlEdit_var(t *testing.T) {
	y := `
name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
y = "name.first";
z = "xyz";
yamlEdit(x, y, z);
`
	o := `name:
  first: xyz
  last: def
`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Yaml Edit with var argument",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"x": {
						dataType: "STRING",
						name:     "x",
						value:    o,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["x"]; ok {
				m := map[string]valHolder{
					"x": d,
				}
				if diff := compare(tt.fields.values, m); !diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleYamlEdit_id_assignment(t *testing.T) {
	y := `name:
  first: abc
  last: def
`
	d := "x = `" + y + "`" + `;
y = x;
`
	o := `name:
  first: abc
  last: def
`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "id to id assignment",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"x": {
						dataType: "STRING",
						name:     "x",
						value:    o,
					},
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    o,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if diff := compare(tt.fields.values, r.Values()); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handlekubectlget(t *testing.T) {
	d := `x = kubectl get -n dev cm/test-cm;
z = "metadata.name";
y = jsonSelect(x, z);
if !x {
  log("missing");
}
`
	a := `items.#(metadata.labels.app\.kubernetes\.io/name=="argocd-server").metadata.name`
	e := `x = kubectl get -n devtroncd po;
z = jsonSelect(x, `
	e = e + "`" + a + "`);"
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl get",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    "test-cm",
					},
				},
			},
			args: args{},
		},
		{
			name: "kubectl get multi",
			fields: fields{
				input: e,
				values: map[string]valHolder{
					"y": {
						dataType: "STRING",
						name:     "y",
						value:    "test-cm",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			if d, ok := r.values["y"]; ok {
				m := map[string]valHolder{
					"y": d,
				}
				if diff := compare(tt.fields.values, m); diff {
					t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
				}
			} else {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handlekubectlapply(t *testing.T) {
	a := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: dev
  labels:
    app.kubernetes.io/instance: my-app
data:
  name: abc
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: dev
  labels:
    app.kubernetes.io/instance: my-app
data:
  name: def
`
	d := "a = `" + a + "`" + `;
x = kubectl apply a;
`
	u := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
update:
  data:
    name: xyz
`
	e := "a = `" + a + "`;\n" + "u = `" + u + "`;" + `
x = kubectl apply a -u u
`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl apply",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"x": {
						dataType: BOOLEAN,
						name:     "x",
						value:    true,
					},
				},
			},
			args: args{},
		},
		{
			name: "kubectl apply with edit",
			fields: fields{
				input: e,
				values: map[string]valHolder{
					"x": {
						dataType: BOOLEAN,
						name:     "x",
						value:    true,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			x := r.values["x"]
			m := map[string]valHolder{
				"x": x,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handlekubectlpatch(t *testing.T) {
	d := `a = kubectl patch -n dev cm test-cm --type "application/merge-patch+json" -p '{"data":{"age":"36"}}';
b = kubectl get -n dev cm test-cm;
c = jsonSelect(b, "data.age");`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl patch",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"c": {
						dataType: STRING,
						name:     "c",
						value:    "36",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			c := r.values["c"]
			m := map[string]valHolder{
				"c": c,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handlekubectldelete(t *testing.T) {
	d := `a = kubectl delete -n dev cm test-cm test-cm2;`
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl delete",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"a": {
						dataType: BOOLEAN,
						name:     "a",
						value:    true,
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			a := r.values["a"]
			m := map[string]valHolder{
				"a": a,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_integration(t *testing.T) {
	a := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: dev
  labels:
    app.kubernetes.io/instance: my-app
data:
  name: abc
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: dev
  labels:
    app.kubernetes.io/instance: my-app
data:
  name: def
`
	d := "a = `" + a + "`" + `;
x = kubectl apply -n dev a;
k = "kind";
z = "metadata.name";
o1 = yamlSelect(a, k, 0) + "/" + yamlSelect(a, z, 0);
k2 = yamlSelect(a, k, 1);
n2 = yamlSelect(a, z, 1);
o2 = k2 + "/" + n2;
age = "36";
pla = '{"data":{"age":"' + age + '"}}';
pa = kubectl patch -n dev k2 n2 --type "application/merge-patch+json" -p pla;
fo = kubectl get -n dev o1 o2;
selector = 'items.#(metadata.name="'+n2+'").data.age';
age = jsonSelect(fo, selector);
`
	//programmers.#(lastName="Hunter").firstName
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl integration",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"age": {
						dataType: "STRING",
						name:     "age",
						value:    "36",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			age := r.values["age"]
			m := map[string]valHolder{
				"age": age,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_temp(t *testing.T) {
	//	base64DecoderPrefix := `#!/bin/bash
	//echo -n "`
	//	base64DecoderSuffix := `" | base64 -d`
	data, err := ioutil.ReadFile("/Users/pghildiy/Documents/devtronCode/installation-yamls/devtron-installation-script/installation-script")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", data)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "kubectl temp",
			fields: fields{
				input: string(data),
				values: map[string]valHolder{
					"age": {
						dataType: "STRING",
						name:     "age",
						value:    "36",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			age := r.values["age"]
			m := map[string]valHolder{
				"age": age,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func TestKlangListener_handleShellScript(t *testing.T) {
	d := `
#!/bin/bash
echo 'hello' | base64`
	d = "a = shellScript `" + d + "`;"
	fmt.Printf("%s", d)
	type fields struct {
		input  string
		values map[string]valHolder
	}
	type args struct {
		ctx *parser2.AssignmentContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "shell script",
			fields: fields{
				input: d,
				values: map[string]valHolder{
					"a": {
						dataType: STRING,
						name:     "a",
						value:    "hello\n",
					},
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setup(tt.fields.input)
			a := r.values["a"]
			m := map[string]valHolder{
				"a": a,
			}
			if diff := compare(tt.fields.values, m); !diff {
				t.Errorf("expected %+v, found %+v\n", tt.fields.values, r.Values())
			}
		})
	}
}

func setup(input string) *KlangListener {
	is := antlr.NewInputStream(input)
	lexer := parser2.NewKlangLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser2.NewKlangParser(stream)
	p.BuildParseTrees = true
	mapper := NewMapperFactory()
	r := NewKlangListener(mapper)
	antlr.ParseTreeWalkerDefault.Walk(r, p.Parse())
	return r
}

func compare(first, second map[string]valHolder) bool {
	return checkFirstInSecond(first, second) && checkFirstInSecond(second, first)
}

func checkFirstInSecond(first, second map[string]valHolder) bool {
	for k, fv := range first {
		if sv, ok := second[k]; ok {
			ft := fmt.Sprintf("%T", fv.value)
			st := fmt.Sprintf("%T", sv.value)
			r := ft != st || fv.dataType != sv.dataType || fv.value != sv.value
			if r {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

//func compareInterface(first, second interface{}) bool {
//	ft := fmt.Sprintf("%T", first)
//	st := fmt.Sprintf("%T", second)
//	if ft != st {
//		return false
//	}
//	switch first.(type) {
//	case int64:
//		return first.(int64) == second.(int64)
//	case float64:
//		return first.(float64) == second.(float64)
//	case bool:
//		return first.(bool) == second.(bool)
//	case string:
//		return first.(string) == second.(string)
//	default:
//		return false
//	}
//}

func Test_handleJsonDelete(t *testing.T) {
	jsonList := `{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}]}`
	json := `{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`
	pattern := "data.school"
	o := unstructured.Unstructured{}
	o.UnmarshalJSON([]byte(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`))
	resourceKey := kube.GetResourceKey(&o)
	filter := (&resourceKey).String()
	type args struct {
		data    string
		filter  string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want valHolder
	}{
		{
			name: "delete in list with filter and pattern",
			args: args{
				data:    jsonList,
				filter:  filter,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}]}`),
		},
		{
			name: "delete in list with filter only",
			args: args{
				data:   jsonList,
				filter: filter,
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school":"abc"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "delete in list with pattern only",
			args: args{
				data:    jsonList,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "delete in object with filter and pattern",
			args: args{
				data:    json,
				filter:  filter,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}`),
		},
		{
			name: "delete in object with pattern only",
			args: args{
				data:    json,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleKubeJsonDelete(tt.args.data, tt.args.filter, tt.args.pattern)
			if got.dataType != tt.want.dataType {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			var gotstruct map[string]interface{}
			var wantStruct map[string]interface{}
			err := json2.Unmarshal([]byte(got.value.(string)), &gotstruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			err = json2.Unmarshal([]byte(tt.want.value.(string)), &wantStruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotstruct, wantStruct) {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleKubeYamlDelete(t *testing.T) {
	ymlList := `
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: service
  metadata:
    name: abc
    namespace: abc
  data:
    school: abc
- apiVersion: v1
  kind: service
  metadata:
    name: def
    namespace: abc
  data:
    school: def`
	//outYmlist := `{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school":"abc"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`
	ymls := `
apiVersion: v1
kind: service
metadata:
  name: abc
  namespace: abc
data:
  school: abc
---
apiVersion: v1
kind: service
metadata:
  name: def
  namespace: abc
data:
  school: def
`
	yml := `
apiVersion: v1
kind: service
metadata:
  name: def
  namespace: abc
data:
  school: def`
	pattern := "data.school"
	o := unstructured.Unstructured{}
	o.UnmarshalJSON([]byte(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`))
	resourceKey := kube.GetResourceKey(&o)
	filter := (&resourceKey).String()
	type args struct {
		data    string
		filter  string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want valHolder
	}{
		{
			name: "delete in list with filter and pattern",
			args: args{
				data:    ymlList,
				filter:  filter,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}]}`),
		},
		{
			name: "delete in list with filter only",
			args: args{
				data:   ymlList,
				filter: filter,
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school":"abc"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "delete in list with pattern only",
			args: args{
				data:    ymlList,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "delete in yamls with filter and pattern",
			args: args{
				data:    ymls,
				filter:  filter,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}]}`),
		},
		{
			name: "delete in yamls with filter only",
			args: args{
				data:   ymls,
				filter: filter,
			},
			want: newStringValHolder(`{"apiVersion":"v1","data":{"school":"abc"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}}`),
		},
		{
			name: "delete in yamls with pattern only",
			args: args{
				data:    ymls,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "delete in object with filter and pattern",
			args: args{
				data:    yml,
				filter:  filter,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}`),
		},
		{
			name: "delete in object with pattern only",
			args: args{
				data:    yml,
				pattern: pattern,
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleKubeYamlDelete(tt.args.data, tt.args.filter, tt.args.pattern)
			if got.dataType != tt.want.dataType {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			oy, err := yaml.YAMLToJSON([]byte(got.value.(string)))
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			yamls := strings.Split(got.value.(string), yamlSeperator)
			if len(yamls) > 1 {
				var jsons []string
				for _, y := range yamls {
					j, err := yaml.YAMLToJSON([]byte(y))
					if err != nil {
						t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
					}
					jsons = append(jsons, string(j))
				}
				oj := strings.Join(jsons, ",")
				oy = []byte(`{"apiVersion": "v1",    "items": [` + oj + `], "kind": "List"}`)
			}

			var gotstruct map[string]interface{}
			var wantStruct map[string]interface{}
			err = json2.Unmarshal(oy, &gotstruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			err = json2.Unmarshal([]byte(tt.want.value.(string)), &wantStruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			diff := cmp.Diff(gotstruct, wantStruct)
			fmt.Printf("%+v\n", diff)
			if !cmp.Equal(gotstruct, wantStruct) {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", gotstruct, wantStruct)
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleKubeJsonEdit(t *testing.T) {
	jsonList := `{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}]}`
	json := `{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`
	pattern := "data.school"
	o := unstructured.Unstructured{}
	o.UnmarshalJSON([]byte(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`))
	resourceKey := kube.GetResourceKey(&o)
	filter := (&resourceKey).String()
	type args struct {
		data    string
		filter  string
		pattern string
		value   interface{}
	}
	tests := []struct {
		name string
		args args
		want valHolder
	}{
		{
			name: "edit in list with filter and pattern",
			args: args{
				data:    jsonList,
				filter:  filter,
				pattern: pattern,
				value:   "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}]}`),
		},
		{
			name: "edit in list with pattern only",
			args: args{
				data:    jsonList,
				pattern: pattern,
				value:   "ghi",
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school": "ghi"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}],"kind":"List"}`),
		},
		{
			name: "edit in object with filter and pattern",
			args: args{
				data:    json,
				filter:  filter,
				pattern: pattern,
				value:   "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}`),
		},
		{
			name: "edit in object with pattern only",
			args: args{
				data:    json,
				pattern: pattern,
				value:   "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleKubeJsonEdit(tt.args.data, tt.args.filter, tt.args.pattern, tt.args.value)
			if got.dataType != tt.want.dataType {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			var gotstruct map[string]interface{}
			var wantStruct map[string]interface{}
			err := json2.Unmarshal([]byte(got.value.(string)), &gotstruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			err = json2.Unmarshal([]byte(tt.want.value.(string)), &wantStruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotstruct, wantStruct) {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleKubeYamlEdit(t *testing.T) {
	ymlList := `
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: service
  metadata:
    name: abc
    namespace: abc
  data:
    school: abc
- apiVersion: v1
  kind: service
  metadata:
    name: def
    namespace: abc
  data:
    school: def`
	//outYmlist := `{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school":"abc"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`
	ymls := `
apiVersion: v1
kind: service
metadata:
  name: abc
  namespace: abc
data:
  school: abc
---
apiVersion: v1
kind: service
metadata:
  name: def
  namespace: abc
data:
  school: def
`
	yml := `
apiVersion: v1
kind: service
metadata:
  name: def
  namespace: abc
data:
  school: def`
	pattern := "data.school"
	o := unstructured.Unstructured{}
	o.UnmarshalJSON([]byte(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "def"}}`))
	resourceKey := kube.GetResourceKey(&o)
	filter := (&resourceKey).String()
	type args struct {
		data    string
		filter  string
		pattern string
		value   interface{}
	}
	tests := []struct {
		name string
		args args
		want valHolder
	}{
		{
			name: "edit in list with filter and pattern",
			args: args{
				data:    ymlList,
				filter:  filter,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}]}`),
		},
		{
			name: "edit in list with pattern only",
			args: args{
				data:    ymlList,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school": "ghi"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{"school": "ghi"},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "edit in yamls with filter and pattern",
			args: args{
				data:    ymls,
				filter:  filter,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "List", "items":[{"apiVersion": "v1", "kind": "service", "metadata": {"name": "abc", "namespace": "abc"}, "data": {"school": "abc"}}, {"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}]}`),
		},
		{
			name: "edit in yamls with pattern only",
			args: args{
				data:    ymls,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion":"v1","items":[{"apiVersion":"v1","data":{"school": "ghi"},"kind":"service","metadata":{"name":"abc","namespace":"abc"}},{"apiVersion":"v1","data":{"school": "ghi"},"kind":"service","metadata":{"name":"def","namespace":"abc"}}],"kind":"List"}`),
		},
		{
			name: "edit in object with filter and pattern",
			args: args{
				data:    yml,
				filter:  filter,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}`),
		},
		{
			name: "edit in object with pattern only",
			args: args{
				data:    yml,
				pattern: pattern,
				value: "ghi",
			},
			want: newStringValHolder(`{"apiVersion": "v1", "kind": "service", "metadata": {"name": "def", "namespace": "abc"}, "data": {"school": "ghi"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleKubeYamlEdit(tt.args.data, tt.args.filter, tt.args.pattern, tt.args.value)
			if got.dataType != tt.want.dataType {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			oy, err := yaml.YAMLToJSON([]byte(got.value.(string)))
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			yamls := strings.Split(got.value.(string), yamlSeperator)
			if len(yamls) > 1 {
				var jsons []string
				for _, y := range yamls {
					j, err := yaml.YAMLToJSON([]byte(y))
					if err != nil {
						t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
					}
					jsons = append(jsons, string(j))
				}
				oj := strings.Join(jsons, ",")
				oy = []byte(`{"apiVersion": "v1",    "items": [` + oj + `], "kind": "List"}`)
			}

			var gotstruct map[string]interface{}
			var wantStruct map[string]interface{}
			err = json2.Unmarshal(oy, &gotstruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			err = json2.Unmarshal([]byte(tt.want.value.(string)), &wantStruct)
			if err != nil {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
			diff := cmp.Diff(gotstruct, wantStruct)
			fmt.Printf("%+v\n", diff)
			if !cmp.Equal(gotstruct, wantStruct) {
				t.Errorf("handleKubeJsonDelete() = %v, want %v", gotstruct, wantStruct)
				t.Errorf("handleKubeJsonDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}