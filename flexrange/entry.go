package flexrange

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"tools/utils"

	gojsonq "github.com/thedevsaddam/gojsonq/v2"
)

var MAPPER = map[string]ExtendDataInt{}

type EntryFormatter struct {
	Eq    string
	Range string
}

type ExtendDataInt interface {
	Copy() utils.CopyAble
	String() string

	MarshalJSON() (b []byte, err error)
	UnmarshalJSON(b []byte) error
	//Left() ExtendDataInt
	//Right() ExtendDataInt
	//AddRight(e ExtendDataInt)
}

type ExtendData struct {
	// 数据类型，需要与 MAPPER 相结合
	// 比如: flexrange.MAPPER["NextHop"] = &NextHop{}
	Type string
	// Property 用于保持最为原始的 RawMessage，用于反序列号
	Property json.RawMessage
	// 从 Property 中反序列化出来的数据结构，数据结构的类型来自于 MAPPER 中的保存
	Data ExtendDataInt
}

func (ext ExtendData) Copy() utils.CopyAble {
	d := ExtendData{}
	d.Type = ext.Type
	d.Data = ext.Data.Copy().(ExtendDataInt)
	d.Property = make([]byte, len(ext.Property))
	copy(d.Property, ext.Property)

	return &d
}

func (ext *ExtendData) String() string {
	b, _ := json.Marshal(ext)
	return string(b)
}

func (ext *ExtendData) UnmarshalJSON(b []byte) error {
	type ed struct {
		Type     string
		Property json.RawMessage
	}
	d := ed{}
	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}
	//fmt.Printf("d.Property = %+v\n", d.Property)

	if _, ok := MAPPER[d.Type]; !ok {
		return fmt.Errorf("unknown data type: '%s'", d.Type)
	}

	data := MAPPER[d.Type]

	ext.Type = d.Type
	ext.Property = make([]byte, len(d.Property))
	copy(ext.Property, d.Property)
	//fmt.Printf("ext.Property = %+v\n", string(ext.Property))

	dataType := reflect.TypeOf(data)

	newInstancePtr := reflect.New(dataType.Elem()).Interface().(ExtendDataInt)
	_ = json.Unmarshal(ext.Property, newInstancePtr)

	ext.Data = newInstancePtr
	return nil
}

type matchstate int

const (
	NONE matchstate = iota
	PART
	FULL
)

func (ms matchstate) String() string {
	return [...]string{"NONE", "PART", "FULL"}[ms]
}

type EntryInt interface {
	Copy() utils.CopyAble
	String() string
	Compare(other EntryInt) position
	SubRemain(other EntryInt) (ms matchstate, left *[]EntryInt, sub *[]EntryInt, remain *[]EntryInt)
	MatchResult(other *EntryList) (match []EntryInt, unmatch []EntryInt)
	Low() *big.Int
	High() *big.Int
	Count() *big.Int
	Data() *ExtendData
	CopyData() *ExtendData
	SetData(*ExtendData)
	//MarshalJSON() (b []byte, err error)
	//UnmarshalJSON(b []byte) error
	WithStrFunc(s func() string)
	WarpperWithDataRange(size uint32, base *big.Int) DataRangeInf
}

type EntryList struct {
	list    []EntryInt
	size    uint32
	base    *big.Int
	strFunc func() string
}

type EntryListInt interface {
	Copy() utils.CopyAble
	String() string
}

func NewEntryList(size uint32, base *big.Int) *EntryList {
	return &EntryList{
		[]EntryInt{},
		size,
		utils.CopyInt(base),
		nil,
	}
}

func NewEntryListFromList(size uint32, base *big.Int, list []EntryInt) *EntryList {
	r := &EntryList{
		[]EntryInt{},
		size,
		utils.CopyInt(base),
		nil,
	}

	for _, e := range list {
		r.list = append(r.list, e.Copy().(EntryInt))
	}
	return r
}

func (el EntryList) Jsonq() *gojsonq.JSONQ {
	text := el.String()
	// jsonq: gojsonq.New().FromString(text),
	return gojsonq.New().FromString(text)
}

func (el EntryList) MarshalJSON() (b []byte, err error) {
	type entrylist struct {
		List []EntryInt
		Size uint32
		Base *big.Int
		//StrFunc func() string
	}

	e := entrylist{
		List: el.list,
		Size: el.size,
		Base: el.base,
		//StrFunc: el.strFunc,
	}
	return json.Marshal(&e)
}

func (el *EntryList) UnmarshalJSON(b []byte) error {
	type entrylist struct {
		//目前只支持Entry
		List []*Entry
		Size uint32
		Base *big.Int
		//不支持strFunc
		//StrFunc func() string
	}
	var e entrylist

	err := json.Unmarshal(b, &e)
	//fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}

	for _, l := range e.List {
		el.list = append(el.list, l)
	}
	el.size = e.Size
	el.base = e.Base
	//el.strFunc = e.StrFunc

	return nil
}

func (el *EntryList) WithStrFunc(s func() string) {
	el.strFunc = s
}

func (el *EntryList) Len() int {
	return len(el.list)
}

func (el *EntryList) Copy() utils.CopyAble {
	ec := NewEntryList(el.Size(), el.Base())
	for it := el.Iterator(); it.HasNext(); {
		_, e := it.Next()
		ec.PushEntry(e.Copy().(EntryInt))
	}
	return ec
}

func (el *EntryList) String() string {
	if el.strFunc == nil {
		a := []string{}
		for it := el.Iterator(); it.HasNext(); {
			_, e := it.Next()
			a = append(a, fmt.Sprintf("%s", e))
		}

		return "[" + strings.Join(a, ",") + "]"
	} else {
		return el.strFunc()
	}
}

type EntryListIterator struct {
	p     *EntryList
	index int
}

func (el *EntryList) Iterator() *EntryListIterator {
	return &EntryListIterator{
		el,
		0,
	}
}

func (it *EntryListIterator) HasNext() bool {
	return it.index < len(it.p.list)
}

func (it *EntryListIterator) Next() (int, EntryInt) {
	e := it.p.list[it.index]
	index := it.index
	it.index++
	return index, e
}

func (it *EntryListIterator) Update(p *EntryList) {
	it.p = p
}

func (el *EntryList) SetBase(b *big.Int) {
	el.base = b
}

func (el *EntryList) Base() *big.Int {
	if el.base == nil {
		return big.NewInt(0)
	}
	return el.base
}

func (el *EntryList) SetSize(s uint32) {
	el.size = s
}

func (el *EntryList) Size() uint32 {
	return el.size
}

func (el *EntryList) MaxValue() *big.Int {
	m := big.NewInt(1).Lsh(big.NewInt(1), uint(el.Size()))
	m = m.Sub(m, big.NewInt(1))
	return m
}

func (el *EntryList) Remove(r EntryInt) (ok bool) {
	for i, e := range el.list {
		if e.Compare(r) == Equal {
			if i == len(el.list)-1 {
				el.list = el.list[0:i]
			} else {
				el.list = append(el.list[0:i], el.list[i+1:]...)
			}
			ok = true
			return
		}
	}

	return
}

func (el *EntryList) MatchResult(other DataRangeInf) (ok bool, err error, match *EntryList, unmatch *EntryList) {
	//unmatch表示el在匹配other以后，还剩余的部分
	match = NewEntryList(other.Size(), other.Base())
	unmatch = NewEntryList(other.Size(), other.Base())
	if el.Size() != other.Size() {
		err = fmt.Errorf("el.Size() = %d, other.Size() = %d", el.Size(), other.Size())
		return
	}

	if el.Base().Cmp(other.Base()) != 0 {
		err = fmt.Errorf("el.Base() = %d, other.Base() = %d", el.Base(), other.Base())
		return
	}

	for it := other.Iterator(); it.HasNext(); {
		_, o := it.Next()
		m, u := o.(EntryInt).MatchResult(el)
		for _, e := range m {
			match.PushEntry(e)
		}

		for _, e := range u {
			unmatch.PushEntry(e)
		}
	}
	ok = true
	return
}

func (el *EntryList) PushString(low string, high string, addition interface{}) (bool, error) {
	l := new(big.Int)
	fmt.Sscan(low, l)
	h := new(big.Int)
	fmt.Sscan(high, h)

	return el.Push(l, h, addition)
}

func (el *EntryList) PushEntry(e EntryInt) (bool, error) {
	el.list = append(el.list, e)
	return true, nil
}

func (el *EntryList) Push(low *big.Int, high *big.Int, addition interface{}) (bool, error) {
	if low.Cmp(high) > 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if low.Cmp(el.Base()) < 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if high.Cmp(el.MaxValue()) > 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}
	var tmp *Entry
	if addition == nil {
		tmp = &Entry{low: utils.CopyInt(low), high: utils.CopyInt(high), data: nil, strFunc: nil}
	} else {
		tmp = &Entry{low: utils.CopyInt(low), high: utils.CopyInt(high), data: addition.(utils.CopyAble).Copy().(*ExtendData), strFunc: nil}
	}
	el.list = append(el.list, tmp)
	return true, nil
}

type Entry struct {
	low  *big.Int
	high *big.Int
	//data interface{}
	//data ExtendDataInt
	data    *ExtendData
	strFunc func() string
}

func NewEntry(low *big.Int, high *big.Int, addition *ExtendData) (*Entry, error) {
	if high.Cmp(low) < 0 {
		return nil, errors.New(fmt.Sprintf("high: %d, low: %d, addition: %T", high, low, addition))
	}

	//if addition == nil {
	//} else {
	//_, ok := addition.(ExtendData)
	//if !ok {
	//return nil, errors.New(fmt.Sprintf("high: %d, low: %d, addition: %T", high, low, addition))
	//}
	//}

	if addition == nil {
		return &Entry{
			low:     utils.CopyInt(low),
			high:    utils.CopyInt(high),
			data:    nil,
			strFunc: nil,
		}, nil

	} else {
		return &Entry{
			low:     utils.CopyInt(low),
			high:    utils.CopyInt(high),
			data:    addition.Copy().(*ExtendData),
			strFunc: nil,
		}, nil
	}
}

func NewEntryFromString(low string, high string, addition *ExtendData) (*Entry, error) {
	l := new(big.Int)
	fmt.Sscan(low, l)
	h := new(big.Int)
	fmt.Sscan(high, h)
	return NewEntry(l, h, addition)
}

func (e Entry) MarshalJSON() (b []byte, err error) {
	type entry struct {
		Low  *big.Int
		High *big.Int
		//Data []byte
		Data *ExtendData
		//Data interface{}
	}
	et := entry{
		Low:  e.low,
		High: e.high,
		Data: e.data,
	}

	//copy(et.Data, b)

	b, err = json.Marshal(&et)
	return
}

func (e *Entry) UnmarshalJSON(b []byte) error {
	type entry struct {
		Low  *big.Int
		High *big.Int
		Data *ExtendData
	}

	var et entry
	err := json.Unmarshal(b, &et)
	if err != nil {
		return err
	}
	e.low = et.Low
	e.high = et.High
	e.data = et.Data

	return nil
}

func (e Entry) Low() *big.Int {
	return e.low
}

func (e Entry) High() *big.Int {
	return e.high
}

func (e Entry) Count() *big.Int {
	var z big.Int
	z.Sub(e.High(), e.Low())
	z.Add(&z, big.NewInt(1))

	return &z
}

//func (e Entry) Data() interface{} {
func (e Entry) Data() *ExtendData {
	return e.data
}

func (e Entry) CopyData() *ExtendData {
	if e.data == nil {
		return nil
	} else {
		return e.data.Copy().(*ExtendData)
	}
}

func (e *Entry) WarpperWithDataRange(size uint32, base *big.Int) DataRangeInf {
	r := NewDataRange(size, base)
	ok, err := r.PushEntry(e.Copy().(*Entry))
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("unknow error")
	}
	return r
}

func (e *Entry) SetData(d *ExtendData) {
	if d != nil {
		//_, ok := d.(*ExtendData)
		//if !ok {
		//err := fmt.Errorf("SetData failed, need implenment ExtendDataInt Interface")
		//panic(err)
		//}
		//fmt.Printf("d = %+v\n", d.Copy())
		e.data = d.Copy().(*ExtendData)
	} else {
		e.data = nil
	}
}

func (e Entry) Copy() utils.CopyAble {
	l := new(big.Int)
	l.Set(e.Low())
	h := new(big.Int)
	h.Set(e.High())

	//var d interface{}
	//if e.data != nil {
	//d = e.data.(utils.CopyAble).Copy()
	//}

	return &Entry{low: l, high: h, data: e.CopyData(), strFunc: nil}
}

//func (e Entry) Copy() *Entry {
//l := new(big.Int)
//l.Set(e.Low)
//h := new(big.Int)
//h.Set(e.High)

//result, _ := NewEntry(l, h)
//return result
//}

func (e Entry) String() string {
	if e.strFunc == nil {
		return fmt.Sprintf("Low: %d, High: %d, Data: %s", e.Low(), e.High(), e.data)
	} else {
		return e.strFunc()
	}
}

func (e *Entry) WithStrFunc(s func() string) {
	e.strFunc = s
}

func (e Entry) Format(format EntryFormatter) string {
	if e.Low().Cmp(e.High()) == 0 {
		return fmt.Sprintf(format.Eq, e.Low())
	} else {
		return fmt.Sprintf(format.Range, e.Low(), e.High())
	}
}

func (e Entry) MatchResult(other *EntryList) (match []EntryInt, unmatch []EntryInt) {
	//unmatch表示e,在匹配了other以后，还剩余未匹配的部分
	unmatch = []EntryInt{}
	match = []EntryInt{}

	ot := other.Copy().(*EntryList)

	var change bool = false
	for it := ot.Iterator(); it.HasNext(); {
		_, o := it.Next()
		ms, left, sub, _ := e.SubRemain(o)
		for _, s := range *sub {
			match = append(match, s)
		}
		if len(*left) == 0 {
			change = true
			return
		} else {
			//fmt.Printf("left = %+v\n", left)
			if ms != NONE {
				change = true
				if len(*left) == 1 {
					tot := other.Copy().(*EntryList)
					m, u := (*left)[0].(*Entry).MatchResult(tot)
					//fmt.Printf("u = %+v\n", u)
					unmatch = append(unmatch, u...)
					match = append(match, m...)
				} else {
					tot := other.Copy().(*EntryList)
					m, u := (*left)[0].(*Entry).MatchResult(tot)
					//fmt.Printf("u = %+v\n", u)
					unmatch = append(unmatch, u...)
					match = append(match, m...)
					tot = other.Copy().(*EntryList)
					m, u = (*left)[1].(*Entry).MatchResult(tot)
					//fmt.Printf("u = %+v\n", u)
					unmatch = append(unmatch, u...)
					match = append(match, m...)
				}
				return
			}
		}
	}
	if !change {
		unmatch = append(unmatch, e.Copy().(EntryInt))
	}
	return

}

func (e Entry) SubRemain(other EntryInt) (ms matchstate, left *[]EntryInt, sub *[]EntryInt, remain *[]EntryInt) {
	//left是e剩余部分
	//sub是共同部分
	//remain是other剩余部分
	ms = NONE
	pos := e.Compare(other)
	ec := e.Copy().(EntryInt)

	sub = &[]EntryInt{}
	remain = &[]EntryInt{}
	left = &[]EntryInt{}

	switch pos {
	case Equal:
		ms = FULL
		*sub = append(*sub, other.Copy().(EntryInt))
		return
	case Left:
		ms = NONE
		*left = append(*left, ec)
		*remain = append(*remain, other.Copy().(EntryInt))
		return
	case LeftConnectRight:
		ms = NONE
		*left = append(*left, ec)
		*remain = append(*remain, other.Copy().(EntryInt))
		return
	case LeftContainRight:
		ms = PART
		if ec.Low().Cmp(other.Low()) == 0 {
			e1 := &Entry{
				low:     utils.AddInt(utils.CopyInt(other.High()), 1),
				high:    utils.CopyInt(ec.High()),
				data:    ec.CopyData(),
				strFunc: nil,
				//ec.Data().(utils.CopyAble).Copy(),
			}
			sub_e1 := other.Copy().(EntryInt)
			*sub = append(*sub, sub_e1)
			*left = append(*left, EntryInt(e1))
			return
		} else if ec.High().Cmp(other.High()) == 0 {
			e1 := &Entry{
				low:     utils.CopyInt(ec.Low()),
				high:    utils.AddInt(utils.CopyInt(other.Low()), -1),
				data:    ec.CopyData(),
				strFunc: nil,
				//ec.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}
			sub_e1 := other.Copy().(EntryInt)
			*sub = append(*sub, sub_e1)
			*left = append(*left, e1)
			return
		} else {
			e1 := &Entry{
				low:  utils.CopyInt(ec.Low()),
				high: utils.AddInt(utils.CopyInt(other.Low()), -1),
				//ec.Data().(utils.CopyAble).Copy(),
				data:    ec.CopyData(),
				strFunc: nil,
				//ec.Data(),
			}
			e2 := &Entry{
				low:     utils.AddInt(utils.CopyInt(other.High()), 1),
				high:    utils.CopyInt(ec.High()),
				data:    ec.CopyData(),
				strFunc: nil,
				//ec.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}

			sub_e1 := other.Copy().(EntryInt)
			*sub = append(*sub, sub_e1)
			*left = append(*left, EntryInt(e1), EntryInt(e2))
			return
		}
	case LeftOverlapRight:
		ms = PART
		e1 := &Entry{
			low:     utils.AddInt(utils.CopyInt(other.High()), 1),
			high:    utils.CopyInt(ec.High()),
			data:    ec.CopyData(),
			strFunc: nil,
			//ec.Data().(utils.CopyAble).Copy(),
			//ec.Data(),
		}

		sub_e1 := &Entry{
			low:     utils.CopyInt(ec.Low()),
			high:    utils.CopyInt(other.High()),
			data:    other.CopyData(),
			strFunc: nil,
			//other.Data().(utils.CopyAble).Copy(),
			//ec.Data(),
		}

		right_e1 := &Entry{
			low:     utils.CopyInt(other.Low()),
			high:    utils.AddInt(utils.CopyInt(ec.Low()), -1),
			data:    other.CopyData(),
			strFunc: nil,
			//other.Data().(utils.CopyAble).Copy(),
			//ec.Data(),
		}

		*sub = append(*sub, sub_e1)
		*left = append(*left, e1)
		*remain = append(*remain, right_e1)
		return
	case Right:
		ms = NONE
		*remain = append(*remain, other.Copy().(EntryInt))
		*left = append(*left, ec)
		return
	case RightConnectLeft:
		ms = NONE
		*remain = append(*remain, other.Copy().(EntryInt))
		*left = append(*left, ec)
		return
	case RightContainLeft:
		ms = FULL
		ecc := ec.(utils.CopyAble).Copy().(EntryInt)
		ecc.SetData(other.CopyData())
		*sub = append(*sub, ecc)
		if ec.Low().Cmp(other.Low()) == 0 {
			e1 := &Entry{
				low:     utils.AddInt(utils.CopyInt(ec.High()), 1),
				high:    utils.CopyInt(other.High()),
				data:    other.CopyData(),
				strFunc: nil,
				//other.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}
			*remain = append(*remain, e1)
			return
		} else if ec.High().Cmp(other.High()) == 0 {
			e1 := &Entry{
				low:     utils.CopyInt(other.Low()),
				high:    utils.AddInt(utils.CopyInt(ec.Low()), -1),
				data:    other.CopyData(),
				strFunc: nil,
				//other.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}
			*remain = append(*remain, e1)
			return
		} else {
			e1 := &Entry{
				low:     utils.CopyInt(other.Low()),
				high:    utils.AddInt(utils.CopyInt(ec.Low()), -1),
				data:    other.CopyData(),
				strFunc: nil,
				//other.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}
			e2 := &Entry{
				low:     utils.AddInt(utils.CopyInt(ec.High()), 1),
				high:    utils.CopyInt(other.High()),
				data:    other.CopyData(),
				strFunc: nil,
				//other.Data().(utils.CopyAble).Copy(),
				//ec.Data(),
			}
			*remain = append(*remain, e1, e2)
			return
		}

	case RightOverlapLeft:
		ms = PART
		e1 := &Entry{
			low:     utils.CopyInt(ec.Low()),
			high:    utils.AddInt(utils.CopyInt(other.Low()), -1),
			data:    ec.CopyData(),
			strFunc: nil,
			//ec.Data().(utils.CopyAble).Copy(),
		}

		sub_e1 := &Entry{
			low:     utils.CopyInt(other.Low()),
			high:    utils.CopyInt(ec.High()),
			data:    other.CopyData(),
			strFunc: nil,
			//other.Data().(utils.CopyAble).Copy(),
			//ec.Data(),
		}
		right_e1 := &Entry{
			low:     utils.AddInt(utils.CopyInt(ec.High()), 1),
			high:    utils.CopyInt(other.High()),
			data:    other.CopyData(),
			strFunc: nil,
			//other.Data().(utils.CopyAble).Copy(),
			//ec.Data(),
		}
		*sub = append(*sub, sub_e1)
		*left = append(*left, e1)
		*remain = append(*remain, right_e1)
		return
	}
	return
}

func (e Entry) Compare(other EntryInt) position {
	//if e.High == other.Low-1 {
	olow := new(big.Int)
	olow.Set(other.Low())
	if olow := olow.Sub(olow, big.NewInt(1)); e.High().Cmp(olow) == 0 {
		return LeftConnectRight
	}

	//if other.High == e.Low-1 {
	elow := new(big.Int)
	elow.Set(e.Low())
	if elow := elow.Sub(elow, big.NewInt(1)); other.High().Cmp(elow) == 0 {
		return RightConnectLeft
	}

	//if e.Low > other.High {
	if e.Low().Cmp(other.High()) > 0 {
		return Right
	}
	//if e.High < other.Low {
	if e.High().Cmp(other.Low()) < 0 {
		return Left
	}
	//if e.Low == other.Low && e.High == other.High {
	if e.Low().Cmp(other.Low()) == 0 && e.High().Cmp(other.High()) == 0 {
		return Equal
	}

	//if e.Low <= other.Low && e.High >= other.High {
	if !(e.Low().Cmp(other.Low()) > 0) && !(e.High().Cmp(other.High()) < 0) {
		return LeftContainRight
	}
	//if e.Low >= other.Low && e.High <= other.High {
	if !(e.Low().Cmp(other.Low()) < 0) && !(e.High().Cmp(other.High()) > 0) {
		return RightContainLeft
	}

	//if e.High < other.High {
	if e.High().Cmp(other.High()) < 0 {
		return RightOverlapLeft
	}

	//if other.High < e.High {
	if other.High().Cmp(e.High()) < 0 {
		return LeftOverlapRight
	}

	return CompareError

}
