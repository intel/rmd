package util

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestNewBitmapUnion(t *testing.T) {
	b, _ := NewBitmap(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitmap([]string{"6-9"})
	r := b.Or(m)
	if r.Bits[0] != 0x1FFF || r.Bits[2] != 0xe00000 {
		t.Errorf("The union should be : 0xE00000,00000000,00001FFF, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}
}

func TestNewBitmap(t *testing.T) {
	b, _ := NewBitmap(96, "3df00cfff00ffafff")
	wants := []int{0xffafff, 0xdf00cfff, 0x3}
	for i, v := range wants {
		if v != b.Bits[i] {
			t.Errorf("The bitmap of index %d should be: 0x%x, but it is: 0x%x",
				i, v, b.Bits[i])
		}
	}

	b, _ = NewBitmap("3df00cfff00ffafff")
	wants = []int{0xffafff, 0xdf00cfff, 0x3}
	for i, v := range wants {
		if v != b.Bits[i] {
			t.Errorf("The bitmap of index %d should be: 0x%x, but it is: 0x%x",
				i, v, b.Bits[i])
		}
	}
}

func TestNewBitmapIntersection(t *testing.T) {
	minlen := 64
	b, _ := NewBitmap(88, []string{"0-7,9-12,32-50,85-87"})
	m, _ := NewBitmap(minlen, []string{"6-9,32-48"})
	r := b.And(m)
	// r.Bits[0]
	len := len(r.Bits)
	if len != 2 {
		t.Error("The length of intersection of bit maps should be %d, but get %d.",
			minlen/32, len)
	}
	if r.Bits[0] != 0x2C0 || r.Bits[1] != 0x1FFFF {
		t.Errorf("The intersection should be : 0x00001FFFF,000002C0, now it is 0x%x,%08x",
			r.Bits[1], r.Bits[0])
	}
}

func TestNewBitmapDifference(t *testing.T) {
	b, _ := NewBitmap(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitmap(64, []string{"6-9"})
	r := b.Xor(m)
	if r.Bits[0] != 0x1d3f || r.Bits[2] != 0xe00000 {
		t.Errorf("The difference should be : 0xE00000,00000000,00001d3f, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}
}

func TestNewBitmapAsymmetricDiff(t *testing.T) {
	minlen := 64
	b, _ := NewBitmap(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitmap(minlen, []string{"6-9"})
	r := b.Axor(m)
	if r.Bits[0] != 0x1c3f || r.Bits[2] != 0xe00000 {
		t.Errorf("The asymmetric difference should be : 0xE00000,00000000,00001c3f, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}

	r = m.Axor(b)
	len := len(r.Bits)
	if len != 2 {
		t.Error("The length of intersection of bit maps should be %d, but get %d.",
			minlen/32, len)
	}
	if r.Bits[0] != 0x100 || r.Bits[1] != 0x0 {
		t.Errorf("The asymmetric difference should be : 0x0,00000100, now it is 0x%x,%08x",
			r.Bits[1], r.Bits[0])
	}
}

func TestBitmapToString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitmap(88, map_list)
	str := b.ToString()
	want := "bffffc,1f000000,00000366"
	if want != str {
		t.Errorf("The value should be '%s', but get '%s'", want, str)
	}

	b, _ = NewBitmap(24, "7FF")
	str = b.ToString()
	want = "7ff"
	if want != str {
		t.Errorf("The value should be '%s', but get '%s'", want, str)
	}
}

func TestBitmapToBinString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitmap(88, map_list)
	str := b.ToBinString()
	want := "101111111111111111111100,00011111000000000000000000000000,00000000000000000000001101100110"
	if want != str {
		t.Errorf("The value should be '%s', but get '%s'", want, str)
	}
}

func TestBitmapToBinStrings(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitmap(88, map_list)
	ss := b.ToBinStrings()

	if len(ss) != 12 {
		t.Error("The length of bit maps string sliece should be %d, but get %d.",
			12, len(ss))
	}
}

func TestBitmapMaxConnectiveBits(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitmap(88, map_list)
	r := b.MaxConnectiveBits()
	want := 0x3FFFFC
	if want != r.Bits[2] {
		t.Errorf("The value should be '%d', but get '%d'", want, r.Bits[2])
	}

	map_list = []string{"1"}
	b, _ = NewBitmap(24, map_list)
	r = b.MaxConnectiveBits()
	want = 0x2
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%d'", want, r.Bits[0])
	}
}

func TestBitmapGetConnectiveBits(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	// 101111111111111111111100,00011111000000000000000000000000,00000000000000000000001101100110
	b, _ := NewBitmap(88, map_list)
	r := b.GetConnectiveBits(10, 10, false)
	want := 0x3FF0
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(3, 4, false)
	want = 0xe0000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(1, 3, false)
	want = 0x100000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(1, 0, false)
	want = 0x800000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	/********************* True **************************************/
	r = b.GetConnectiveBits(2, 3, true)
	want = 0x60
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%x'", want, r.Bits[0])
	}

	r = b.GetConnectiveBits(1, 3, true)
	want = 0x20
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%x'", want, r.Bits[0])
	}
}

func TestGenCPUResStringSimple(t *testing.T) {
	map_list := []string{"0-7"}
	s, e := GenCPUResString(map_list, 88)
	if e != nil {
		t.Errorf("Get CpuResString error: %v", e)
	}

	fmt.Println(s)
	// Output:
	// 0,0,ff
}

func TestGenCPUResString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	map_bin := []string{"1101100110",
		"00011111000000000000000000000000", "101111111111111111111100"}

	s, e := GenCPUResString(map_list, 88)
	if e != nil {
		t.Errorf("Get CpuResString error: %v", e)
	}

	cpus := strings.Split(s, ",")
	len := len(cpus)
	if len != 3 {
		t.Error("Get Wrong cpus map string.")
	}

	for i, v := range map_bin {
		v1, _ := strconv.ParseInt(v, 2, 64)
		v2, _ := strconv.ParseInt(cpus[len-1-i], 16, 64)
		if v1 != v2 {
			t.Errorf("The bitmap of index %d should be: %s, but it is: %s",
				i, v, fmt.Sprintf("%b", v2))
		}
	}
}

func TestGenCPUResStringOutofRange(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-88,^86,^61-65,1024"}
	_, e := GenCPUResString(map_list, 88)
	if e != nil {
		reason := fmt.Sprintf(
			"The biggest index %d is not less than the bit map length %d", 1024, 88)
		es := fmt.Sprintf("%v", e)
		if reason == es {
			t.Log(es)
		} else {
			t.Errorf("Get CpuResString error: %v", e)
		}
	} else {
		t.Errorf("Get CpuResString should failed.")
	}

}

func TestGenCPUResStringWithWrongExpression(t *testing.T) {
	map_list := []string{"abc1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	_, e := GenCPUResString(map_list, 88)
	if e != nil {
		reason := "wrong expression"
		es := fmt.Sprintf("%v", e)
		if strings.Contains(es, reason) {
			t.Log(es)
		} else {
			t.Errorf("Get CpuResString error: %v", e)
		}
	}
}

func TestString2data(t *testing.T) {
	hex_datas := []uint{0xffffff0f, 0xf1, 0xff2fff}
	datas, _ := string2data("ff2fff,f1,ffffff0f")
	for i, v := range datas {
		if v != hex_datas[i] {
			t.Errorf("Parser error, the %d element should be: 0x%x, but get: 0x%x. \n",
				i, hex_datas[i], v)
		} else {
			fmt.Printf("Parser %d element, get: 0x%x. \n", i, v)
		}
	}
	hex_datas = []uint{0x00ffafff, 0xdf00cfff, 0x3}
	datas, _ = string2data("3df00cfff00ffafff")
	for i, v := range datas {
		if v != hex_datas[i] {
			t.Errorf("Parser error, the %d element should be: 0x%x, but get: 0x%x.\n",
				i, hex_datas[i], v)
		} else {
			fmt.Printf("Parser %d element, get: 0x%x. \n", i, v)
		}
	}
}

func TestIsEmptyBitMap(t *testing.T) {

	cpus := "000000,00000000,00000000"
	b, _ := NewBitmap(88, cpus)
	if !b.IsEmpty() {
		t.Errorf("Parser error, the %s element is empty bit map\n", cpus)
	}

	cpus = "000000,00000000,00000001"
	b, _ = NewBitmap(88, cpus)
	if b.IsEmpty() {
		t.Errorf("Parser error, the %s element is not empty bit map\n", cpus)
	}
}

// Below cases are for resctrl schemata
func TestBitmap(t *testing.T) {
	mask0 := ""
	bm0, _ := NewBitmap(11, mask0)
	if !bm0.IsEmpty() {
		t.Errorf("Wrong mask, it should be empty")
	}

	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)
	if bm1.ToString() != "7ff" {
		t.Errorf("Wrong mask, it should be 7ff")
	}
}

func TestBitmapXor(t *testing.T) {
	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)

	mask2 := "f3"
	bm2, _ := NewBitmap(11, mask2)
	bm1 = bm1.Xor(bm2)
	if bm1.ToString() != "70c" {
		t.Errorf("Wrong mask, 7ff xor f3 must be 70c")
	}
}

func TestBitmapAnd(t *testing.T) {
	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)

	mask2 := "f3"
	bm2, _ := NewBitmap(11, mask2)
	bm1 = bm1.And(bm2)
	if bm1.ToString() != "f3" {
		t.Errorf("Wrong mask, 7ff xor f3 must be f3")
	}
}

func TestBitmapOr(t *testing.T) {
	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)

	mask2 := "f800"
	bm2, _ := NewBitmap(16, mask2)
	bm1 = bm1.Or(bm2)

	if bm1.ToString() != "ffff" {
		t.Errorf("Wrong mask, 7ff xor f800 must be ffff")
	}
}

func TestBitmapMaxConnectiveBitsForSchemata(t *testing.T) {
	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)

	cb := bm1.MaxConnectiveBits()
	if cb.ToString() != "7ff" {
		t.Errorf("Wrong mask, max connective bits must be 7ff")
	}

	mask2 := "7f0f"
	bm2, _ := NewBitmap(16, mask2)

	cb = bm2.MaxConnectiveBits()
	if cb.ToString() != "7f00" {
		t.Errorf("Wrong mask, max connective bits must be 7f00")
	}
}

func TestBitmapGetConnectiveBitsForSchemata(t *testing.T) {
	mask1 := "f03"
	bm1, _ := NewBitmap(12, mask1)

	if bm1.GetConnectiveBits(3, 0, true).ToString() != "700" {
		t.Errorf("Wrong mask, get 3 connective bits from f03 low bit offset = 0 must be 700")
	}

	if bm1.GetConnectiveBits(3, 1, true).ToString() != "700" {
		t.Errorf("Wrong mask, get 3 connective bits from f03 low bit offset = 1 must be 700")
	}

	if bm1.GetConnectiveBits(2, 0, true).ToString() != "3" {
		t.Errorf("Wrong mask, get 2 connective bits from f03 low bit offset = 0 must be 3")
	}

	if bm1.GetConnectiveBits(2, 1, true).ToString() != "300" {
		t.Errorf("Wrong mask, get 2 connective bits from f03 low bit offset = 1 must be 300")
	}

	if bm1.GetConnectiveBits(1, 1, true).ToString() != "2" {
		t.Errorf("Wrong mask, get 1 connective bits from f03 low bit offset = 1 must be 1")
	}

	if bm1.GetConnectiveBits(3, 0, false).ToString() != "e00" {
		t.Errorf("Wrong mask, get 3 connective bits from f03 high bit offset = 0 must be e00")
	}

	if bm1.GetConnectiveBits(3, 1, false).ToString() != "700" {
		t.Errorf("Wrong mask, get 3 connective bits from f03 high bit offset = 1 must be 700")
	}

	if bm1.GetConnectiveBits(2, 0, false).ToString() != "c00" {
		t.Errorf("Wrong mask, get 2 connective bits from f03 high bit offset = 1 must be c00")
	}

	if bm1.GetConnectiveBits(2, 1, false).ToString() != "600" {
		t.Errorf("Wrong mask, get 2 connective bits from f03 high bit offset = 1 must be 600")
	}
}

func TestBitmapGetConnectiveBitsEmpty(t *testing.T) {
	mask1 := "7ff"
	bm1, _ := NewBitmap(11, mask1)

	if !bm1.GetConnectiveBits(12, 1, true).IsEmpty() {
		t.Errorf("Wrong mask, get 12 connective bits from 7ff low bit offset = 1 must be empty")
	}

	if !bm1.GetConnectiveBits(12, 0, false).IsEmpty() {
		t.Errorf("Wrong mask, get 12 connective bits from 7ff high bit offset = 0 must be empty")
	}

	if !bm1.GetConnectiveBits(12, 1, false).IsEmpty() {
		t.Errorf("Wrong mask, get 12 connective bits from 7ff high bit offset = 1 must be empty")
	}

	mask2 := "101"
	bm2, _ := NewBitmap(9, mask2)

	if !bm2.GetConnectiveBits(2, 100, false).IsEmpty() {
		t.Errorf("Wrong mask, get 2 connective bits from 101 high bit offset = 100 must be empty")
	}

	if !bm2.GetConnectiveBits(2, 0, false).IsEmpty() {
		t.Errorf("Wrong mask, get 2 connective bits from 101 high bit offset = 0 must be empty")
	}

}

func TestBitmapMaximum(t *testing.T) {
	mask1 := ""
	bm1, _ := NewBitmap(mask1)
	if bm1.Maximum() != 0 {
		t.Errorf("The maximum bit of empty Bitmap must be 0")
	}

	mask1 = "101"
	bm1, _ = NewBitmap(mask1)
	if bm1.Maximum() != 9 {
		t.Errorf("The maximum bit of 101 must be 9")
	}

	mask1 = "1fff,f0000001"
	bm1, _ = NewBitmap(mask1)
	if bm1.Maximum() != 45 {
		t.Errorf("The maximum bit of 1fff,f0000001 must be 45")
	}

	mask1 = "30,00000000,f0000001"
	bm1, _ = NewBitmap(mask1)
	if bm1.Maximum() != 70 {
		t.Errorf("The maximum bit of 30,00000000,f0000001 must be 70")
	}

}

func TestBitMapToHumanString(t *testing.T) {
	bm, _ := NewBitmap("f00000,00000010,f011000c")
	if bm.ToHumanString() != "2-3,16,20,28-31,36,84-87" {
		t.Errorf("Humman string for f00000,00000010,f011000c should be 2-3,16,20,28-31,36,84-87")
	}

	bm, _ = NewBitmap("8fffff,00000000,fffffff3")
	if bm.ToHumanString() != "0-1,4-31,64-83,87" {
		t.Errorf("Humman string for 8fffff,00000000,fffffff3 should be 0-1,4-31,64-83,87")
	}

	bm, _ = NewBitmap("ffffff,ffffffff,ffffffff")
	if bm.ToHumanString() != "0-87" {
		t.Errorf("Humman string for ffffff,ffffffff,fffffff should be 0-87")
	}

	bm, _ = NewBitmap("1f")
	if bm.ToHumanString() != "0-4" {
		t.Errorf("Humman string for 1f should be 0-4")
	}

	bm, _ = NewBitmap("")
	if bm.ToHumanString() != "" {
		t.Errorf("Humman string for `` should be ``")
	}
}
