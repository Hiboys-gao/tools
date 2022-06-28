package flexrange

import (
	"encoding/json"
	"fmt"
	"math/big"
	"tools/utils"
)

type position int

const (
	Left position = iota
	Right
	Equal
	LeftOverlapRight
	RightOverlapLeft
	LeftConnectRight
	RightConnectLeft
	LeftContainRight
	RightContainLeft
	CompareError
)

//type Result struct {
//match   DataRangeInf
//unmatch DataRangeInf
//}

func (p position) String() string {
	return [...]string{"Left", "Right", "Equal", "LeftOverlapRight", "RightOverlapLeft", "LeftConnectRight", "RightConnectLeft", "LeftContainRight", "RightContainLeft", "CompareError"}[p]
}

type DataRangeInf interface {
	Match(other DataRangeInf) bool
	//MatchResult(other DataRangeInf) Result
	Same(other DataRangeInf) bool
	Iterator() *Iterator
	Base() *big.Int
	Count() *big.Int
	SetBase(b *big.Int)
	SetSize(s uint32)
	Size() uint32
	//SetList(el []*Entry)
	SetList(el []EntryInt)
	//List() []*Entry
	List() []EntryInt
	Empty() bool
	MaxValue() *big.Int
	//Copy() DataRangeInf
	Copy() utils.CopyAble
	//Insert(index int, e *Entry) (bool, error)
	Delete(index int) (bool, error)
	//PushString(low string, high string, addition interface{}) (bool, error)
	PushString(low string, high string, addition *ExtendData) (bool, error)
	PushEntry(e EntryInt) (bool, error)
	//Push(low *big.Int, high *big.Int, addition interface{}) (bool, error)
	Push(low *big.Int, high *big.Int, addition *ExtendData) (bool, error)
	RemoveString(low string, high string) (DataRangeInf, error)
	Remove(low *big.Int, high *big.Int) (DataRangeInf, error)
	Add(other DataRangeInf) DataRangeInf
	Sub(other DataRangeInf) (DataRangeInf, error)
	ElementStrFunc(func() string)
	WithStrFunc(func() string)
	//Sub(other DataRangeInf) (DataRangeInf, DataRangeInf, error)
}

type DataRange struct {
	L []EntryInt
	//L []interface{}
	//utils.BaseList
	//list []*Entry
	size    uint32
	base    *big.Int
	StrFunc func() string
}

func (dr DataRange) MarshalJSON() (b []byte, err error) {
	type datarange struct {
		L []EntryInt
		//utils.BaseList
		Size uint32
		Base *big.Int
	}

	d := datarange{
		L:    dr.L,
		Size: dr.size,
		Base: dr.base,
	}
	//d.L = dr.L

	return json.Marshal(&d)
}

func (dr *DataRange) UnmarshalJSON(b []byte) error {
	type datarange struct {
		//期待机构优化，目前支持Entry的Unmarshal，暂时够用
		L    []*Entry
		Size uint32
		Base *big.Int
	}

	var d datarange
	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}

	for _, i := range d.L {
		dr.L = append(dr.L, i)
	}
	dr.size = d.Size
	dr.SetBase(d.Base)
	return nil
}

//func (dr *DataRange) SetList(el []*Entry) {
func (dr *DataRange) SetList(el []EntryInt) {
	dr.L = el
}

func (dr *DataRange) SetBase(b *big.Int) {
	dr.base = b
}

func (d *DataRange) Base() *big.Int {
	if d.base == nil {
		return big.NewInt(0)
	}
	return d.base
}

func (d *DataRange) Count() *big.Int {
	var z big.Int
	for it := d.Iterator(); it.HasNext(); {
		_, e := it.Next()
		z.Add(&z, e.Count())
	}

	return &z
}

func (dr *DataRange) SetSize(s uint32) {
	dr.size = s
}

func (d *DataRange) Size() uint32 {
	return d.size
}

//func (d *DataRange) List() []*Entry {
func (d *DataRange) List() []EntryInt {
	return d.L
}

func (d *DataRange) Empty() bool {
	if len(d.L) == 0 {
		return true
	}

	return false
}

func (d *DataRange) String() string {
	if d.StrFunc == nil {
		var ls string
		for it := d.Iterator(); it.HasNext(); {
			_, e := it.Next()
			//fmt.Printf("reflect.TypeOf(e) = %+v\n", reflect.TypeOf(e))
			if ls == "" {
				ls = fmt.Sprintf("%s", e)
			} else {
				ls = fmt.Sprintf("%s, %s", ls, e)
			}
		}
		//return fmt.Sprintf("List: %s, Size: %d, Base: %d", fmt.Sprintf("%+v", d.List()), d.Size(), d.Base())
		return fmt.Sprintf("List:[%s], Size:%d, Base:%d", ls, d.Size(), d.Base())
	} else {
		return d.StrFunc()
	}
}

func (d *DataRange) Iterator() *Iterator {
	return &Iterator{
		dr:    d,
		index: 0,
	}
}

type Iterator struct {
	dr    *DataRange
	index int
}

func (i *Iterator) HasNext() bool {
	return i.index < len(i.dr.List())
}

func (i *Iterator) Update(dr *DataRange) {
	i.dr = dr
}

func (i *Iterator) Next() (index int, v EntryInt) {
	//func (i *Iterator) Next() (index int, v interface{}) {
	//v = i.dr.List()[i.index]
	v = i.dr.List()[i.index].(EntryInt)
	index = i.index
	i.index++
	return index, v
}

func (i *Iterator) Delete(index int) EntryInt {
	if index < 0 || index > len(i.dr.List()) {
		return nil
	}
	if index < i.index {
		i.index = i.index - 1
	}

	d := i.dr.List()[index].(EntryInt)

	i.dr.Delete(i.index)
	return d
}

func (i *Iterator) Add(index int, v EntryInt) {
	if index < 0 || index > len(i.dr.List())+1 {
		return
	}
	i.dr.Insert(i.index, v)
	if index < i.index {
		i.index = i.index - 1
	}
}

func NewDataRange(size uint32, base *big.Int) *DataRange {
	if size > 128 {
		return nil
	}
	d := &DataRange{
		L:       []EntryInt{},
		size:    size,
		base:    base,
		StrFunc: nil,
	}

	return d
}

func (dr *DataRange) Delete(index int) (bool, error) {
	if index < 0 || index > len(dr.List()) {
		return false, fmt.Errorf("index: %d, len(dr.List): %d", index, len(dr.List()))
	}

	if index == len(dr.List())-1 {
		dr.L = dr.List()[0:index]
	} else {
		dr.L = append(dr.List()[0:index], dr.List()[index+1:]...)
	}

	return true, nil
}

//func (dr *DataRange) Insert(index int, s interface{}) (bool, error) {
func (dr *DataRange) Insert(index int, s EntryInt) (bool, error) {
	if index < 0 || index > len(dr.List()) {
		return false, fmt.Errorf("index: %d, len(r.List): %d", index, len(dr.List()))
	}
	if index == len(dr.List())-1 {
		dr.L = append(dr.List()[0:index+1], dr.List()[index:]...)
		dr.List()[index] = s
	} else {
		dr.L = append(dr.List()[0:index+1], dr.List()[index:]...)
		dr.List()[index] = s
	}
	return true, nil
}

func (d *DataRange) Same(other DataRangeInf) bool {
	return d.Match(other) && other.Match(d)
}

//func (d DataRange) MatchDataRange(other *DataRange) bool {
func (d *DataRange) Match(other DataRangeInf) bool {
	//for _i, i in other.next():
	//match = False
	//for _j, j in self.next():
	//test = DataRange.test_range(i, j)
	//if test == 'EQUAL' or test == 'RIGHT_CONTAIN_LEFT':
	//match = True
	//break
	//if not match:
	//return False
	//return True

	for it := other.Iterator(); it.HasNext(); {
		match := false
		_, o := it.Next()

		for it2 := d.Iterator(); it2.HasNext(); {
			_, t := it2.Next()
			test := o.(EntryInt).Compare(t.(EntryInt))
			if test == Equal || test == RightContainLeft {
				match = true
				break
			}
		}
		if !match {
			return false
		}

	}
	return true

}

//func (d *DataRange) MatchResult(other DataRangeInf) Result {
//oc := other.Copy().(DataRangeInf)

//for it := oc.Iterator(); it.HasNext(); {
//_, o := it.Next()

//for it2 := d.Iterator(); it2.HasNext(); {
//_, t := it2.Next()
//}
//if !match {
//return false
//}

//}
//return true

//}

func (d *DataRange) MaxValue() *big.Int {
	m := big.NewInt(1).Lsh(big.NewInt(1), uint(d.Size()))
	m = m.Sub(m, big.NewInt(1))
	return m
}

func (d *DataRange) ElementStrFunc(strFunc func() string) {
	for _, e := range d.List() {
		e.WithStrFunc(strFunc)
	}
}

func (d *DataRange) WithStrFunc(strFunc func() string) {
	d.StrFunc = strFunc
}

//func (d *DataRange) Copy() DataRangeInf {
//result := NewDataRange(d.Size(), d.Base())

//for _, e := range d.List() {
//result.L = append(result.List(), e.(*Entry).Copy())
//}
//return result
//}

func (d *DataRange) Copy() utils.CopyAble {
	result := NewDataRange(d.Size(), d.Base())

	for _, e := range d.List() {
		result.L = append(result.List(), e.(EntryInt).Copy().(EntryInt))
	}
	return result
}

func (r *DataRange) Print() {
	for _, d := range r.List() {
		fmt.Printf("%+v\n", d)
	}
}

//func (r *DataRange) Insert(index int, e *Entry) (bool, error) {
//if index < 0 || index > len(r.List()) {
//return false, fmt.Errorf("index: %d, len(r.List): %d", index, len(r.List()))
//}
//if index == len(r.List())-1 {
//r.L = append(r.List()[0:index+1], r.List()[index:]...)
//r.List()[index] = e
//} else {
//r.L = append(r.List()[0:index+1], r.List()[index:]...)
//r.List()[index] = e
//}
//return true, nil
//}

//func (r *DataRange) Delete(index int) (bool, error) {
//if index < 0 || index > len(r.List()) {
//return false, fmt.Errorf("index: %d, len(r.List): %d", index, len(r.List()))
//}

//if index == len(r.List())-1 {
//r.L = r.List()[0:index]
//} else {
//r.L = append(r.List()[0:index], r.List()[index+1:]...)
//}

//return true, nil
//}

//func (d *DataRange) PushString(low string, high string, addition interface{}) (bool, error) {
func (d *DataRange) PushString(low string, high string, addition *ExtendData) (bool, error) {
	l := new(big.Int)
	fmt.Sscan(low, l)
	h := new(big.Int)
	fmt.Sscan(high, h)
	return d.Push(l, h, addition)
}

func (d *DataRange) PushEntry(e EntryInt) (bool, error) {
	ec := e.Copy().(EntryInt)

	return d.Push(ec.Low(), ec.High(), ec.Data())
}

func (d *DataRange) Push(low *big.Int, high *big.Int, addition *ExtendData) (bool, error) {
	//func (d *DataRange) Push(low *big.Int, high *big.Int, addition interface{}) (bool, error) {
	if low.Cmp(high) > 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if low.Cmp(d.Base()) < 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if high.Cmp(d.MaxValue()) > 0 {
		return false, fmt.Errorf("low: %d, high: %d", low, high)
	}

	if len(d.List()) == 0 {
		d.L = append(d.List(), &Entry{low: low, high: high, data: addition, strFunc: nil})
		return true, nil
	}

	tmp := &Entry{low: low, high: high, data: addition, strFunc: nil}
	for it := d.Iterator(); it.HasNext(); {
		index, cur := it.Next()
		pos := tmp.Compare(cur.(*Entry))
		if pos == Equal {
			break
		} else if pos == Left {
			d.Insert(index, tmp)
			break
		} else if pos == Right {
			if index == len(d.List())-1 {
				d.L = append(d.List(), tmp)
				break
			}
		} else if pos == RightContainLeft {
			break
		} else if pos == LeftContainRight {
			h := new(big.Int)
			h.Set(cur.(*Entry).High())
			d.List()[index].(*Entry).Low().Set(tmp.Low())
			if tmp.High().Cmp(h) == 0 {
				break
			} else if index == len(d.List())-1 {
				d.List()[index].(*Entry).High().Set(tmp.High())
				break
			} else {
				h = h.Add(h, big.NewInt(1))
				tmp.Low().Set(h)
			}

		} else if pos == RightOverlapLeft {
			//self.update(index, 'low', d['low'])
			d.List()[index].(*Entry).Low().Set(tmp.Low())
		} else if pos == LeftOverlapRight {
			//result.push(low=d['low'], high=self.get(index)['high'],
			//addition=self.get(index)['addition'])
			//high = cur['high']
			//self.update(index, 'high', d['low'] - 1)
			//d['low'] = high + 1
			//continue
			if index == len(d.List())-1 {
				d.List()[index].(*Entry).High().Set(tmp.High())
				break
			} else {
				h := new(big.Int)
				h.Set(cur.(*Entry).High())
				h = h.Add(h, big.NewInt(1))
				tmp = &Entry{
					low:     h,
					high:    tmp.high,
					data:    tmp.data,
					strFunc: nil,
				}
			}
		} else if pos == LeftConnectRight {
			//self.update(index, 'low', d['low'])
			d.List()[index].(*Entry).Low().Set(tmp.Low())
		} else if pos == RightConnectLeft {
			//if index == self.len() - 1:
			//self.update(index, 'high', d['high'])
			//break
			//else:
			//d = {'low': cur['high'] + 1, 'high': d['high'], 'addition': d['addition']}
			if index == len(d.List())-1 {
				d.List()[index].(*Entry).High().Set(tmp.High())
				break
			} else {
				h := new(big.Int)
				h.Set(cur.(*Entry).High())
				h = h.Add(h, big.NewInt(1))
				tmp = &Entry{
					low:     h,
					high:    tmp.high,
					data:    tmp.data,
					strFunc: nil,
				}

			}
		}

	}
	//for (index, cur) in self.next():
	//if index < self.len() - 1:
	//test = DataRange.test_range(cur, self.get(index + 1))
	//if test == 'LEFT_CONNECT_RIGHT':
	//cur['high'] = self.get(index + 1)['high']
	//self.delete(index + 1)
Loop:
	for it := d.Iterator(); it.HasNext(); {
		index, cur := it.Next()
		if index < len(d.List())-1 {
			pos := cur.(EntryInt).Compare(d.List()[index+1].(EntryInt))
			if pos == LeftConnectRight {
				d.List()[index].(EntryInt).High().Set(d.List()[index+1].(EntryInt).High())
				if index+1 == len(d.List())-1 {
					d.L = d.List()[0 : index+1]
				} else {
					d.L = append(d.List()[0:index+1], d.List()[index+2:]...)
				}

				goto Loop
			}
		}
	}
	return true, nil
}
func (dr *DataRange) RemoveString(low string, high string) (DataRangeInf, error) {
	l := new(big.Int)
	fmt.Sscan(low, l)
	h := new(big.Int)
	fmt.Sscan(high, h)

	return dr.Remove(l, h)

}

func (dr *DataRange) Remove(low *big.Int, high *big.Int) (DataRangeInf, error) {
	if low.Cmp(high) > 0 {
		return nil, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if low.Cmp(dr.Base()) < 0 {
		return nil, fmt.Errorf("low: %d, high: %d", low, high)
	}
	if high.Cmp(dr.MaxValue()) > 0 {
		return nil, fmt.Errorf("low: %d, high: %d", low, high)
	}

	if len(dr.List()) == 0 {
		return nil, nil
	}

	d := &Entry{low: low, high: high, data: nil, strFunc: nil}
	//result用于保持dr中被删除的部分，用于返回
	result := NewDataRange(dr.Size(), dr.Base())
	for it := dr.Iterator(); it.HasNext(); {
		index, cur := it.Next()
		//fmt.Printf("cur = %+v\n", cur)
		//fmt.Printf("d = %+v\n", d)
		pos := d.Compare(cur.(*Entry))
		//fmt.Println("VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV")
		//fmt.Printf("d = %+v\n", d)
		//fmt.Printf("cur = %+v\n", cur)
		//fmt.Printf("pos = %+v\n", pos)
		//fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
		if pos == Equal {
			result.Push(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
			//dr.Delete(index)
			it.Delete(index)
			break
		} else if pos == Left {
			break
		} else if pos == Right {
			continue
		} else if pos == LeftConnectRight {
			break
		} else if pos == RightConnectLeft {
			continue
		} else if pos == LeftContainRight {
			if d.Low().Cmp(cur.(*Entry).Low()) == 0 {
				d.Low().Set(utils.AddInt(cur.(*Entry).High(), 1))
				result.Push(utils.CopyInt(cur.(*Entry).Low()),
					utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
				//dr.Delete(index)
				it.Delete(index)
				continue

			} else {
				result.Push(utils.CopyInt(cur.(*Entry).Low()),
					utils.CopyInt(cur.(*Entry).High()),
					cur.(*Entry).Data())
				h := utils.CopyInt(cur.(*Entry).High())
				//l := CopyInt(cur.Low)

				//dr.Delete(index)
				it.Delete(index)
				if d.High().Cmp(h) == 0 {
					break
				} else {
					d.Low().Set(utils.AddInt(h, 1))
					continue
				}

			}
		} else if pos == RightContainLeft {
			result.Push(utils.CopyInt(d.Low()), utils.CopyInt(d.High()), cur.(*Entry).Data())
			if d.Low().Cmp(cur.(*Entry).Low()) == 0 {
				//fmt.Printf("AddInt(d.High, 1) = %+v\n", AddInt(d.High, 1))
				dr.List()[index].(*Entry).Low().Set(utils.AddInt(d.High(), 1))
				break
			} else if d.High().Cmp(cur.(*Entry).High()) == 0 {
				//fmt.Printf("d.Low = %+v\n", d.Low)
				//fmt.Printf("AddInt(d.Low, -1) = %+v\n", AddInt(d.Low, -1))
				dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
				break
			} else {
				h := utils.CopyInt(cur.(*Entry).High())
				//fmt.Printf("d.Low = %+v\n", d.Low)
				//fmt.Printf("AddInt(d.Low, -1) = %+v\n", AddInt(d.Low, -1))
				dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
				e := &Entry{
					low:     utils.AddInt(d.High(), 1),
					high:    h,
					data:    cur.(*Entry).Data(),
					strFunc: nil}
				//fmt.Printf("index = %+v\n", index)
				//fmt.Printf("dr = %+v\n", dr)
				//dr.L = append(dr.List(), e)
				if index == len(dr.List())-1 {
					dr.L = append(dr.List(), e)
				} else {
					dr.Insert(index+1, e)
				}
				break
			}

		} else if pos == LeftOverlapRight {
			result.Push(utils.CopyInt(d.Low()), utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
			h := utils.CopyInt(cur.(*Entry).High())
			dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
			d.Low().Set(utils.AddInt(h, 1))
			continue
		} else if pos == RightOverlapLeft {
			result.Push(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(d.High()), cur.(*Entry).Data())
			dr.List()[index].(*Entry).Low().Set(utils.AddInt(d.High(), 1))
			break
		}
	}
	return result, nil
}

//func (dr *DataRange) Remove(low *big.Int, high *big.Int) (DataRangeInf, error) {
//if low.Cmp(high) > 0 {
//return nil, fmt.Errorf("low: %d, high: %d", low, high)
//}
//if low.Cmp(dr.Base()) < 0 {
//return nil, fmt.Errorf("low: %d, high: %d", low, high)
//}
//if high.Cmp(dr.MaxValue()) > 0 {
//return nil, fmt.Errorf("low: %d, high: %d", low, high)
//}

//if len(dr.List()) == 0 {
//return nil, nil
//}

//d := &Entry{low, high, nil}
////result用于保持dr中被删除的部分，用于返回
//result := NewDataRange(dr.Size(), dr.Base())
//for it := dr.Iterator(); it.HasNext(); {
//index, cur := it.Next()
//fmt.Printf("cur = %+v\n", cur)
//fmt.Printf("d = %+v\n", d)
//pos := d.Compare(cur.(*Entry))
//if pos == Equal {
//result.Push(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
//dr.Delete(index)
//break
//} else if pos == Left {
//break
//} else if pos == Right {
//continue
//} else if pos == LeftConnectRight {
//break
//} else if pos == RightConnectLeft {
//continue
//} else if pos == LeftContainRight {
//if d.Low().Cmp(cur.(*Entry).Low()) == 0 {
////d['low'] = cur['high'] + 1
////result.push(low=self.get(index)['low'], high=self.get(index)['high'],
////addition=self.get(index)['addition'])
////self.delete(index)

////continue
////c := cur.Copy()
//d.Low().Set(utils.AddInt(cur.(*Entry).High(), 1))
//result.Push(utils.CopyInt(cur.(*Entry).Low()),
//utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
//dr.Delete(index)
//continue

//} else {
////result.push(low=self.get(index)['low'], high=self.get(index)['high'],
////addition=self.get(index)['addition'])
////high = self.get(index)['high']
////low = self.get(index)['low']
////self.delete(index)

////if d['high'] == high:
////break
////else:
////d['low'] = high + 1
////continue
//result.Push(utils.CopyInt(cur.(*Entry).Low()),
//utils.CopyInt(cur.(*Entry).High()),
//cur.(*Entry).Data())
//h := utils.CopyInt(cur.(*Entry).High())
////l := CopyInt(cur.Low)

//dr.Delete(index)
//if d.High().Cmp(h) == 0 {
//break
//} else {
//d.Low().Set(utils.AddInt(h, 1))
//continue
//}

//}
//} else if pos == RightContainLeft {
////result.push(low=d['low'], high=d['high'], addition=self.get(index)['addition'])
////if d['low'] == cur['low']:
////self.update(index, 'low', d['high'] + 1)

////break
////elif d['high'] == cur['high']:
////self.update(index, 'high', d['low'] - 1)

////break
////else:
////high = cur['high']
////self.update(index, 'high', d['low'] - 1)

////self.insert(index + 1, {'low': d['high'] + 1, 'high': high, 'addition':
////self.get(index)['addition']})
////break
//result.Push(utils.CopyInt(d.Low()), utils.CopyInt(d.High()), cur.(*Entry).Data())
//if d.Low().Cmp(cur.(*Entry).Low()) == 0 {
////fmt.Printf("AddInt(d.High, 1) = %+v\n", AddInt(d.High, 1))
//dr.List()[index].(*Entry).Low().Set(utils.AddInt(d.High(), 1))
//break
//} else if d.High().Cmp(cur.(*Entry).High()) == 0 {
////fmt.Printf("d.Low = %+v\n", d.Low)
////fmt.Printf("AddInt(d.Low, -1) = %+v\n", AddInt(d.Low, -1))
//dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
//break
//} else {
//h := utils.CopyInt(cur.(*Entry).High())
////fmt.Printf("d.Low = %+v\n", d.Low)
////fmt.Printf("AddInt(d.Low, -1) = %+v\n", AddInt(d.Low, -1))
//dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
//e := &Entry{
//utils.AddInt(d.High(), 1),
//h,
//cur.(*Entry).Data()}
////fmt.Printf("index = %+v\n", index)
////fmt.Printf("dr = %+v\n", dr)
//dr.L = append(dr.List(), e)
////dr.Insert(index+1, e)
//break
//}

//} else if pos == LeftOverlapRight {
////result.push(low=d['low'], high=self.get(index)['high'],
////addition=self.get(index)['addition'])
////high = cur['high']
////self.update(index, 'high', d['low'] - 1)
////d['low'] = high + 1
////continue
//result.Push(utils.CopyInt(d.Low()), utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
//h := utils.CopyInt(cur.(*Entry).High())
//dr.List()[index].(*Entry).High().Set(utils.AddInt(d.Low(), -1))
//d.Low().Set(utils.AddInt(h, 1))
//continue
//} else if pos == RightOverlapLeft {
////result.push(low=self.get(index)['low'], high=d['high'],
////addition=self.get(index)['addition'])
////self.update(index, 'low', d['high'] + 1)
////break
//result.Push(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(d.High()), cur.(*Entry).Data())
//dr.List()[index].(*Entry).Low().Set(utils.AddInt(d.High(), 1))
//break
//}
//}
//return result, nil
//}

func (dr *DataRange) Add(other DataRangeInf) DataRangeInf {
	if dr.Base().Cmp(other.Base()) != 0 {
		return nil
	}
	if other == nil {
		return nil
	}

	if other.Size() != dr.Size() {
		return nil
	}

	dr_copy := dr.Copy()
	//rm := NewDataRange(dr.Size, dr.Base)

	for it := other.Iterator(); it.HasNext(); {
		_, cur := it.Next()
		_, err := dr_copy.(*DataRange).Push(utils.CopyInt(cur.(*Entry).Low()),
			utils.CopyInt(cur.(*Entry).High()),
			cur.(*Entry).Data())
		if err != nil {
			return nil
		}
	}
	for it := other.Iterator(); it.HasNext(); {
		_, cur := it.Next()
		dr.Push(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()), cur.(*Entry).Data())
	}

	return dr
}

//dr中保持减去other剩余的部分
//返回共同减去的部分
func (dr *DataRange) Sub(other DataRangeInf) (DataRangeInf, error) {
	if dr.Base().Cmp(other.Base()) != 0 {
		return nil, fmt.Errorf("other: %+v", other)
	}
	if other == nil {
		return nil, fmt.Errorf("other: %+v", other)
	}

	if other.Size() != dr.Size() {
		return nil, fmt.Errorf("other.Size: %d, dr.Size: %d", other.Size(), dr.Size())
	}
	dr_copy := dr.Copy()
	rm := NewDataRange(dr.Size(), dr.Base())
	//for it := dr_copy.(*DataRange).Iterator(); it.HasNext(); {
	//_, e := it.Next()
	//fmt.Printf("e = %+v\n", e)
	//}
	for it := other.Iterator(); it.HasNext(); {
		_, cur := it.Next()
		_, err := dr_copy.(*DataRange).Remove(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()))
		if err != nil {
			return nil, err
		}
	}

	for it := other.Iterator(); it.HasNext(); {
		_, cur := it.Next()
		//tmp保持了dr.Remove操作后，被删除的部分
		tmp, err := dr.Remove(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()))
		if err != nil {
			return nil, err
		}
		//将dr中被删除的部分累加起来
		if tmp != nil {
			for it2 := tmp.Iterator(); it2.HasNext(); {
				_, e := it2.Next()
				rm.PushEntry(e.(*Entry))
			}

		}

	}

	//返回dr中被删除掉的部分数据
	return rm, nil
}

//func (dr *DataRange) Sub(other DataRangeInf) (DataRangeInf, error) {
//if dr.Base().Cmp(other.Base()) != 0 {
//return nil, fmt.Errorf("other: %+v", other)
//}
//if other == nil {
//return nil, fmt.Errorf("other: %+v", other)
//}

//if other.Size() != dr.Size() {
//return nil, fmt.Errorf("other.Size: %d, dr.Size: %d", other.Size(), dr.Size())
//}
//dr_copy := dr.Copy()
//rm := NewDataRange(dr.Size(), dr.Base())
//for it := other.Iterator(); it.HasNext(); {
//_, cur := it.Next()
//_, err := dr_copy.(*DataRange).Remove(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()))
//if err != nil {
//return nil, err
//}
//}

//for it := other.Iterator(); it.HasNext(); {
//_, cur := it.Next()
////tmp保持了dr.Remove操作后，被删除的部分
//tmp, err := dr.Remove(utils.CopyInt(cur.(*Entry).Low()), utils.CopyInt(cur.(*Entry).High()))
//if err != nil {
//return nil, err
//}
////将dr中被删除的部分累加起来
//if tmp != nil {
//for it2 := tmp.Iterator(); it2.HasNext(); {
//_, e := it2.Next()
//rm.PushEntry(e.(*Entry))
//}

//}

//}

////返回dr中被删除掉的部分数据
//return rm, nil
//}

func DataRangeCmp(this DataRangeInf, other DataRangeInf) (left DataRangeInf, mid DataRangeInf, right DataRangeInf) {

	if this.Size() != other.Size() {
		return this.Copy().(*DataRange), nil, other.Copy().(*DataRange)
	}

	this_dr := this.Copy().(*DataRange)
	other_dr := other.Copy().(*DataRange)
	this_dr_copy := this_dr.Copy().(*DataRange)
	rm_this, _ := this_dr.Sub(other_dr)
	_, _ = other_dr.Sub(this_dr_copy)

	if len(rm_this.List()) == 0 {
		mid = nil
	} else {
		mid = rm_this
	}

	if len(this_dr.List()) == 0 {
		left = nil
	} else {
		left = this_dr
	}
	if len(other_dr.List()) == 0 {
		right = nil
	} else {
		right = other_dr
	}
	return
}

type DataRangePair struct {
	one DataRangeInf
	two DataRangeInf
}

func NewDataRangePair(one DataRangeInf, two DataRangeInf) *DataRangePair {
	if one.Empty() {
		return nil
	}
	if two.Empty() {
		return nil
	}
	return &DataRangePair{
		one.Copy().(*DataRange),
		two.Copy().(*DataRange),
	}
}

func DataRangePairCmp(this DataRangePair, other DataRangePair) (left *[]DataRangePair, mid *[]DataRangePair, right *[]DataRangePair) {

	left1, mid1, right1 := DataRangeCmp(this.one, other.one)
	left2, mid2, right2 := DataRangeCmp(this.two, other.two)
	if left1 == nil {
		left1 = NewDataRange(this.one.Size(), this.one.Base())
	}
	if left2 == nil {
		left2 = NewDataRange(this.two.Size(), this.two.Base())
	}

	if mid1 == nil {
		mid1 = NewDataRange(this.one.Size(), this.one.Base())
	}
	if mid2 == nil {
		mid2 = NewDataRange(this.two.Size(), this.two.Base())
	}

	if right1 == nil {
		right1 = NewDataRange(other.one.Size(), other.one.Base())
	}
	if right2 == nil {
		right2 = NewDataRange(other.two.Size(), other.two.Base())
	}

	l1 := NewDataRangePair(left1, other.one.Copy().(*DataRange))
	l2 := NewDataRangePair(mid1, left2)
	if l1 == nil && l2 == nil {
		left = nil
	} else if l1 == nil {
		left = &[]DataRangePair{
			*l2,
		}
	} else if l2 == nil {
		left = &[]DataRangePair{
			*l1,
		}
	} else {
		left = &[]DataRangePair{
			*l1,
			*l2,
		}
	}

	r1 := NewDataRangePair(right1, other.two.Copy().(*DataRange))
	r2 := NewDataRangePair(mid1, right2)

	if r1 == nil && r2 == nil {
		right = nil
	} else if r1 == nil {
		right = &[]DataRangePair{
			*r2,
		}
	} else if r2 == nil {
		right = &[]DataRangePair{
			*r1,
		}
	} else {
		right = &[]DataRangePair{
			*r1,
			*r2,
		}
	}

	m1 := NewDataRangePair(mid1, mid2)
	if m1 == nil {
		mid = nil
	} else {
		mid = &[]DataRangePair{
			*m1,
		}
	}

	return

}
