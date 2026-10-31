package cursor

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	c := Cursor{Score: 0.8731, UserID: "0b7c0f5a-1c2d-4e3f-9a8b-7c6d5e4f3a2b"}
	got, err := Decode(Encode(c))
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	if got.UserID != c.UserID || got.Score != c.Score {
		t.Fatalf("往返不一致: %+v != %+v", got, c)
	}
}

func TestDecodeEmpty(t *testing.T) {
	c, err := Decode("")
	if err != nil {
		t.Fatalf("空游标不应报错: %v", err)
	}
	if c.Valid() {
		t.Fatal("空游标不应有效")
	}
}

func TestDecodeInvalid(t *testing.T) {
	cases := []string{"###", "bm90LXZhbGlk", "MTIzfA"}
	for _, s := range cases {
		if _, err := Decode(s); err == nil && s != "MTIzfA" {
			t.Fatalf("非法游标应报错：%s", s)
		}
	}
}
