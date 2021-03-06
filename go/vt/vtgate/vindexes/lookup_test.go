/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vindexes

import (
	"reflect"
	"testing"

	"strings"

	"github.com/youtube/vitess/go/sqltypes"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
)

var lookupUnique Vindex
var lookupNonUnique Vindex

func init() {
	lkpunique, err := CreateVindex("lookup_unique", "lookupUnique", map[string]string{"table": "t", "from": "fromc", "to": "toc"})
	if err != nil {
		panic(err)
	}
	lkpnonunique, err := CreateVindex("lookup", "lookupNonUnique", map[string]string{"table": "t", "from": "fromc", "to": "toc"})
	if err != nil {
		panic(err)
	}

	lookupUnique = lkpunique
	lookupNonUnique = lkpnonunique
}

func TestLookupUniqueCost(t *testing.T) {
	if lookupUnique.Cost() != 10 {
		t.Errorf("Cost(): %d, want 10", lookupUnique.Cost())
	}
}

func TestLookupNonUniqueCost(t *testing.T) {
	if lookupNonUnique.Cost() != 20 {
		t.Errorf("Cost(): %d, want 20", lookupUnique.Cost())
	}
}

func TestLookupUniqueString(t *testing.T) {
	if strings.Compare("lookupUnique", lookupUnique.String()) != 0 {
		t.Errorf("String(): %s, want lookupUnique", lookupUnique.String())
	}
}

func TestLookupNonUniqueString(t *testing.T) {
	if strings.Compare("lookupNonUnique", lookupNonUnique.String()) != 0 {
		t.Errorf("String(): %s, want lookupNonUnique", lookupNonUnique.String())
	}
}

func TestLookupUniqueVerify(t *testing.T) {
	vc := &vcursor{numRows: 1}
	_, err := lookupUnique.Verify(vc, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test")})
	wantQuery := &querypb.BoundQuery{
		Sql: "select fromc from t where ((fromc=:fromc0 and toc=:toc0))",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc0": sqltypes.Int64BindVariable(1),
			"toc0":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}

	//Negative test
	want := "lookup.Verify:length of ids 2 doesn't match length of ksids 1"
	_, err = lookupUnique.Verify(vc, []sqltypes.Value{testVal(1), testVal(2)}, [][]byte{[]byte("test")})
	if err.Error() != want {
		t.Error(err.Error())
	}

	_, err = lookuphashunique.Verify(nil, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test1test23")})
	want = "lookup.Verify: invalid keyspace id: 7465737431746573743233"
	if err.Error() != want {
		t.Error(err)
	}
}

func TestLookupUniqueMap(t *testing.T) {
	vc := &vcursor{}
	_, err := lookupUnique.(Unique).Map(vc, []sqltypes.Value{testVal(2)})
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "select toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(2),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}
}

func TestLookupUniqueCreate(t *testing.T) {
	vc := &vcursor{}
	err := lookupUnique.(Lookup).Create(vc, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test")})
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "insert into t(fromc,toc) values(:fromc0,:toc0)",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc0": sqltypes.Int64BindVariable(1),
			"toc0":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}

	//Negative test
	want := "lookup.Create:length of ids 2 doesn't match length of ksids 1"
	err = lookupUnique.(Lookup).Create(vc, []sqltypes.Value{testVal(1), testVal(2)}, [][]byte{[]byte("test")})
	if err.Error() != want {
		t.Error(err.Error())
	}

	err = lookuphashunique.(Lookup).Create(nil, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test1test23")})
	want = "lookup.Create: invalid keyspace id: 7465737431746573743233"
	if err.Error() != want {
		t.Error(err)
	}
}

func TestLookupUniqueReverse(t *testing.T) {
	_, ok := lookupUnique.(Reversible)
	if ok {
		t.Errorf("lhu.(Reversible): true, want false")
	}
}

func TestLookupUniqueDelete(t *testing.T) {
	vc := &vcursor{}
	err := lookupUnique.(Lookup).Delete(vc, []sqltypes.Value{testVal(1)}, []byte("test"))
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "delete from t where fromc = :fromc and toc = :toc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(1),
			"toc":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}

	//Negative Test
	err = lookuphashunique.(Lookup).Delete(vc, []sqltypes.Value{testVal(1)}, []byte("test1test23"))
	want := "lookup.Delete: invalid keyspace id: 7465737431746573743233"
	if err.Error() != want {
		t.Error(err)
	}
}

func TestLookupNonUniqueVerify(t *testing.T) {
	vc := &vcursor{numRows: 1}
	_, err := lookupNonUnique.Verify(vc, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test")})
	wantQuery := &querypb.BoundQuery{
		Sql: "select fromc from t where ((fromc=:fromc0 and toc=:toc0))",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc0": sqltypes.Int64BindVariable(1),
			"toc0":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}
}

func TestLookupNonUniqueMap(t *testing.T) {
	vc := &vcursor{}
	_, err := lookupNonUnique.(NonUnique).Map(vc, []sqltypes.Value{testVal(2)})
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "select toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(2),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}
}

func TestLookupNonUniqueCreate(t *testing.T) {
	vc := &vcursor{}
	err := lookupNonUnique.(Lookup).Create(vc, []sqltypes.Value{testVal(1)}, [][]byte{[]byte("test")})
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "insert into t(fromc,toc) values(:fromc0,:toc0)",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc0": sqltypes.Int64BindVariable(1),
			"toc0":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}
}

func TestLookupNonUniqueReverse(t *testing.T) {
	_, ok := lookupNonUnique.(Reversible)
	if ok {
		t.Errorf("lhu.(Reversible): true, want false")
	}
}

func TestLookupNonUniqueDelete(t *testing.T) {
	vc := &vcursor{}
	err := lookupNonUnique.(Lookup).Delete(vc, []sqltypes.Value{testVal(1)}, []byte("test"))
	if err != nil {
		t.Error(err)
	}
	wantQuery := &querypb.BoundQuery{
		Sql: "delete from t where fromc = :fromc and toc = :toc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(1),
			"toc":   sqltypes.BytesBindVariable([]byte("test")),
		},
	}
	if !reflect.DeepEqual(vc.bq, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.bq, wantQuery)
	}
}
