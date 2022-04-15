package config

import (
	"reflect"
	"testing"
)

func TestCommandlineConfigParseNext(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		prefix   string
		expectKV *kv
	}{
		// No prefix
		{[]string{"-cn=test-cn"}, "", &kv{"cn", "test-cn"}},
		{[]string{"--cn=test-cn"}, "", &kv{"cn", "test-cn"}},
		{[]string{"-cn", "test-cn"}, "", &kv{"cn", "test-cn"}},
		{[]string{"--cn", "test-cn"}, "", &kv{"cn", "test-cn"}},

		// With prefix
		{[]string{"-someprefix.cn=test-cn"}, "someprefix", &kv{"cn", "test-cn"}},
		{[]string{"--someprefix.cn=test-cn"}, "someprefix", &kv{"cn", "test-cn"}},
		{[]string{"-someprefix.cn", "test-cn"}, "someprefix", &kv{"cn", "test-cn"}},
		{[]string{"--someprefix.cn", "test-cn"}, "someprefix", &kv{"cn", "test-cn"}},

		// Prefix filter
		{[]string{"-cn=test-cn"}, "someprefix", nil},
		{[]string{"-cn", "test-cn"}, "someprefix", nil},
		{[]string{"-otherprefix.cn=test-cn"}, "someprefix", nil},
		{[]string{"-otherprefix.cn", "test-cn"}, "someprefix", nil},

		// Mutiple level key
		{[]string{"-key.alg=rsa"}, "", &kv{"key.alg", "rsa"}},
		{[]string{"-key.size=2048"}, "", &kv{"key.size", "2048"}},

		// Quoted string
		{[]string{"-key.size=\"2048\""}, "", &kv{"key.size", "\"2048\""}},
		{[]string{"-key.size='2048'"}, "", &kv{"key.size", "'2048'"}},
		{[]string{"-key.size", "\"2048\""}, "", &kv{"key.size", "\"2048\""}},
		{[]string{"-key.size", "'2048'"}, "", &kv{"key.size", "'2048'"}},
	} {
		c := NewCommandlineConfig(tc.args, tc.prefix)
		kv, err := c.parseNext()
		if err != nil {
			t.Errorf("args [%v] fail, Parse config error: %v", tc.args, err)
			continue
		}

		if tc.expectKV == nil && kv != nil {
			t.Errorf("args [%v] fail, expectKV==nil but kv=nil\n", tc.args)
			continue
		}
		if tc.expectKV != nil && kv == nil {
			t.Errorf("args [%v] fail, expectKV!=nil but kv==nil\n", tc.args)
			continue
		}

		if tc.expectKV == nil {
			continue
		}

		if kv.key != tc.expectKV.key || kv.val != tc.expectKV.val {
			t.Errorf("args [%v] fail, kv(%v)!=expectKV(%v)\n", tc.args, kv, tc.expectKV)
			continue
		}
	}
}

func TestCommandlineConfigParse(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		expectKVs []*kv
	}{
		{
			[]string{
				"-cn=test-cn",
				"-name=\"test-cn\"",         // Double quote
				"-label='test'",             // Single quote
				"-serial.sid=\"999\"",       // Quoted number
				"-serial.big2=\"1024.123\"", // Quoted float
				"-key.alg=rsa",
				"-key.size=2048",
				"-serial.attr.name=serial1",
				"-serial.big=1024.123",
				"-serial.small=-1024.123",
				"-serial.attr.name=serial2",
			},
			[]*kv{
				// String value
				{"cn", "test-cn"},
				{"key.alg", "rsa"},
				// With quotes
				{"name", "test-cn"},         // Double quoted string, the quotes should be stripped.
				{"label", "test"},           // Single quoted string, the quotes should be stripped
				{"serial.sid", "999"},       // Quoted int is interpreted as string too
				{"serial.big2", "1024.123"}, // Quoted float is interpreted as string too
				// int/float
				{"key.size", int64(2048)},            // int
				{"serial.big", float64(1024.123)},    // float64
				{"serial.small", float64(-1024.123)}, // Nagtive float64
				// Not exists
				{"key.alg_not_exists", nil},
				/// Right priority
				{"serial.attr.name", "serial2"}, // The priority is from last to previous, thus, the last arguments will replace the previous one.
			},
		},
	} {
		c := NewCommandlineConfig(tc.args, "")
		err := c.Parse()
		if err != nil {
			t.Errorf("Test args [%v] fail, Parse config error: %v", tc.args, err)
			continue
		}

		for _, expectKV := range tc.expectKVs {
			v := c.Get(expectKV.key)

			var vv any
			switch expectKV.val.(type) {
			case string:
				vv = v.ToString("")
			case int64:
				vv = v.ToInt(0)
			case bool:
				vv = v.ToBool(false)
			case float64:
				vv = v.ToFloat(0.0)
			case nil:
				// Expect not exists
				// Set vv=nil
				vv = nil
			default:
				t.Logf("Warning: Test args %v, key \"%v\", unsupported value type: %v, you can use type constraint on expectValue\n",
					tc.args, expectKV.key, reflect.TypeOf(expectKV.val))
				continue
			} // endof switch

			if vv != expectKV.val {
				t.Errorf("Test args %v, key \"%v\", value(\"%v\")!=expectValue(\"%v\")",
					tc.args, expectKV.key, vv, expectKV.val)
				continue
			}
		} // endof for _, expectKV {}
	} // endof for _, tc {}
}

/*
func TestCommandlineConfigUnmarshal(t *testing.T) {
	args := []string{
		"-key.alg=rsa",
		"-key.size=2048",
	}
	c := NewCommandlineConfig(args, "")
	err := c.Parse()
	if err != nil {
		t.Fatalf("Test args [%v] fail, Parse config error: %v", args, err)
	}
	v := c.Get("key")
	if v == nil {
		t.Fatalf("\"%v\" not found", "key")
	}

	var kc struct {
		Alg  string `yaml:"alg,omitempty"`
		Size int    `yaml:"size,omitempty"`
	}
	fmt.Println(v)
	err = v.Unmarshal(&kc)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if kc.Alg != "rsa" {
		t.Errorf("kc.Alg(\"%v\")!=expect(\"%v\")", kc.Alg, "rsa")
	}
	if kc.Size != 2048 {
		t.Errorf("kc.Size(\"%v\")!=expect(\"%v\")", kc.Alg, 2048)
	}
}*/
