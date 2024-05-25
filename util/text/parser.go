package text

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/bobappleyard/lync/util/queue"
)

type UnexpectedToken struct {
	Token any
}

func (e *UnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token: %#v", e.Token)
}

type Parser[T, U any] struct {
	root    *symbol
	ruleSet any
}

func NewParser[T, U any](ruleSet any) Parser[T, U] {
	s := &scanner{
		host:      reflect.ValueOf(ruleSet),
		tokenType: reflect.TypeOf(new(T)).Elem(),
		rootType:  reflect.TypeOf(new(U)).Elem(),
		types:     map[reflect.Type]*symbol{},
	}

	root := s.scan()
	return Parser[T, U]{
		root:    root,
		ruleSet: ruleSet,
	}
}

func (p Parser[T, U]) Parse(toks []T) (U, error) {
	var zero U
	tokVals := make([]reflect.Value, len(toks))
	for i, t := range toks {
		tokVals[i] = reflect.ValueOf(t)
	}

	m := &matcher{
		state: [][]item{nil},
		toks:  tokVals,
	}
	if err := m.run(p.root); err != nil {
		return zero, err
	}

	rv, err := m.builder().build(p.root)
	if err != nil {
		return zero, err
	}

	return rv.Interface().(U), nil
}

type symbol struct {
	// this symbol can be empty
	nullable bool

	// if this is a token rule
	tokenType reflect.Type

	// if this is a nonterminal rule
	predictions []*rule
}

type rule struct {
	// symbol this implements
	implements *symbol

	// array of symbols to match
	deps []*symbol

	// the parser host
	host reflect.Value

	// debug: the rule's method name
	name string

	// index of method into host
	index int

	// function to call when building the parse tree
	method func(host reflect.Value, args []reflect.Value) []reflect.Value
}

type scanner struct {
	host      reflect.Value
	tokenType reflect.Type
	rootType  reflect.Type
	types     map[reflect.Type]*symbol
}

func (s *scanner) scan() *symbol {
	s.ensure(s.rootType)
	s.scanMethods(s.host)
	s.markTokenTypes()
	s.markNullableTypes()
	s.fillOutInterfaces()

	return s.types[s.rootType]
}

func (s *scanner) scanMethods(host reflect.Value) {
	hostType := host.Type()
	for i := hostType.NumMethod() - 1; i >= 0; i-- {
		m := hostType.Method(i)
		if !m.IsExported() {
			continue
		}
		deps := make([]*symbol, m.Type.NumIn()-1)
		for i := m.Type.NumIn() - 1; i >= 1; i-- {
			deps[i-1] = s.ensure(m.Type.In(i))
		}
		if m.Type.Out(0).Kind() == reflect.Slice {
			panic("explicit slice rules are not supported")
		}
		produces := s.ensure(m.Type.Out(0))
		produces.predictions = append(produces.predictions, &rule{
			implements: produces,
			deps:       deps,
			host:       host,
			name:       m.Name,
			index:      m.Index,
			method: func(host reflect.Value, args []reflect.Value) []reflect.Value {
				return m.Func.Call(args)
			},
		})
	}
}

func (s *scanner) markTokenTypes() {
	for k, v := range s.types {
		if k.AssignableTo(s.tokenType) {
			v.tokenType = k
			continue
		}
	}
}

func (s *scanner) markNullableTypes() {
	var needsWork queue.Queue[*symbol]
	symUsers := map[*symbol][]*rule{}

	for _, sym := range s.types {
		for _, r := range sym.predictions {
			for _, s := range r.deps {
				symUsers[s] = append(symUsers[s], r)
			}
			if len(r.deps) == 0 {
				sym.nullable = true
				needsWork.Enqueue(sym)
			}
		}
	}

	for needsWork.Ready() {
		next := needsWork.Dequeue()
	nextRule:
		for _, r := range symUsers[next] {
			if r.implements.nullable {
				continue
			}
			for _, s := range r.deps {
				if !s.nullable {
					continue nextRule
				}
			}
			r.implements.nullable = true
			needsWork.Enqueue(r.implements)
		}
	}
}

func (s *scanner) fillOutInterfaces() {
	var itfs []reflect.Type
	for k := range s.types {
		if k.Kind() != reflect.Interface {
			continue
		}
		itfs = append(itfs, k)
	}
	for len(itfs) != 0 {
		s.fillOutInterface(&itfs, itfs[0])
	}
}

func (s *scanner) fillOutInterface(itfs *[]reflect.Type, todo reflect.Type) {
	if !s.needsFilling(itfs, todo) {
		return
	}
	for k, v := range s.types {
		if k == todo {
			continue
		}
		if !k.AssignableTo(todo) {
			continue
		}
		if k.Kind() == reflect.Interface {
			s.fillOutInterface(itfs, k)
		}
		sym := s.types[todo]
		for _, r := range v.predictions {
			sym.predictions = append(sym.predictions, &rule{
				implements: sym,
				deps:       r.deps,
				host:       r.host,
				name:       r.name,
				index:      r.index,
				method:     r.method,
			})
		}
	}
}

func (s *scanner) needsFilling(itfs *[]reflect.Type, todo reflect.Type) bool {
	set := *itfs
	for i, t := range set {
		if t != todo {
			continue
		}
		copy(set[i:], set[i+1:])
		set = set[:len(set)-1]
		*itfs = set
		return true
	}
	return false
}

func (s *scanner) ensure(key reflect.Type) *symbol {
	if v, ok := s.types[key]; ok {
		return v
	}
	v := new(symbol)
	s.types[key] = v
	if key.Kind() == reflect.Slice {
		s.sliceTypeSymbol(v, key)
	} else if m, ok := key.MethodByName("Parser"); ok {
		host := m.Func.Call([]reflect.Value{
			reflect.New(key).Elem(),
		})[0]
		s.scanMethods(host)
	}
	return v
}

func (s *scanner) sliceTypeSymbol(sliceSym *symbol, slice reflect.Type) {
	elem := slice.Elem()
	elemSym := s.ensure(elem)
	sliceSym.predictions = append(sliceSym.predictions, &rule{
		implements: sliceSym,
		deps:       []*symbol{},
		host:       s.host,
		name:       fmt.Sprintf("[]%s(nil)", elem),
		index:      -1,
		method: func(host reflect.Value, args []reflect.Value) []reflect.Value {
			res := reflect.MakeSlice(slice, 0, 0)
			return []reflect.Value{res}
		},
	})
	sliceSym.predictions = append(sliceSym.predictions, &rule{
		implements: sliceSym,
		deps:       []*symbol{sliceSym, elemSym},
		host:       s.host,
		name:       fmt.Sprintf("[]%s(append)", elem),
		index:      -1,
		method: func(host reflect.Value, args []reflect.Value) []reflect.Value {
			res := reflect.Append(args[1], args[2])
			return []reflect.Value{res}
		},
	})
}

type matcher struct {
	state [][]item
	toks  []reflect.Value
	cur   int
}

type item struct {
	// the rule that this item is matching
	rule *rule

	// where in the input this item begins
	position int

	// how far through the rule this item has progressed
	progress int
}

func (p *matcher) run(root *symbol) error {
	p.state = [][]item{nil}
	p.predict(root)
	for _, t := range p.toks {
		p.state = append(p.state, nil)

		p.step(t)
		p.cur++
	}
	p.finalStep()
	return p.matches(root)
}

func (p *matcher) step(tok reflect.Value) {
	for i := 0; i < len(p.state[p.cur]); i++ {
		item := p.state[p.cur][i]
		next, ok := item.nextSymbol()
		if !ok {
			p.complete(item)
			continue
		}
		if next.tokenType != nil {
			if tok.Type().AssignableTo(next.tokenType) {
				p.scan(item)
			}
			continue
		}
		if next.nullable {
			p.advance(item)
		}
		p.predict(next)
	}
}

func (p *matcher) finalStep() {
	for i := 0; i < len(p.state[p.cur]); i++ {
		item := p.state[p.cur][i]
		next, ok := item.nextSymbol()
		if !ok {
			p.complete(item)
			continue
		}
		if next.nullable {
			p.advance(item)
			p.predict(next)
		}
	}
}

func (p *matcher) matches(root *symbol) error {
	if len(p.state[len(p.state)-1]) == 0 {
		for i := range p.state[1:] {
			if len(p.state[i+1]) != 0 {
				continue
			}
			return &UnexpectedToken{
				p.toks[i],
			}
		}
	}
	for _, item := range p.state[len(p.state)-1] {
		if item.rule.implements != root {
			continue
		}
		if item.position != 0 {
			continue
		}
		if _, ok := item.nextSymbol(); ok {
			continue
		}
		return nil
	}
	return io.ErrUnexpectedEOF
}

func (p *matcher) predict(s *symbol) {
	for _, prediction := range s.predictions {
		p.addToCur(item{
			rule:     prediction,
			position: p.cur,
		})
	}
}

func (p *matcher) advance(x item) {
	p.addToCur(x.makeProgress())
}

func (p *matcher) scan(x item) {
	p.addToNext(x.makeProgress())
}

func (p *matcher) complete(x item) {
	for _, y := range p.state[x.position] {
		next, ok := y.nextSymbol()
		if !ok {
			continue
		}
		if next == x.rule.implements {
			p.addToCur(y.makeProgress())
		}
	}
}

func (p *matcher) addToCur(x item) {
	p.addTo(p.cur, x)
}

func (p *matcher) addToNext(x item) {
	p.addTo(p.cur+1, x)
}

func (p *matcher) addTo(pos int, x item) {
	for _, y := range p.state[pos] {
		if x == y {
			return
		}
	}
	p.state[pos] = append(p.state[pos], x)
}

func (x item) complete() bool {
	_, ok := x.nextSymbol()
	return !ok
}

func (x item) nextSymbol() (*symbol, bool) {
	if x.progress == len(x.rule.deps) {
		return nil, false
	}
	return x.rule.deps[x.progress], true
}

func (x item) makeProgress() item {
	return item{
		rule:     x.rule,
		position: x.position,
		progress: x.progress + 1,
	}
}

var (
	ErrFailedMatch    = errors.New("failed to match")
	ErrAmbiguousParse = errors.New("ambiguous parse")
)

type builder struct {
	state [][]item
	seen  []reflect.Value
}

type span struct {
	item     item
	at       int
	value    reflect.Value
	children []span
}

func (p *matcher) builder() *builder {
	flipped := p.flipState()
	for _, s := range flipped {
		sort.Slice(s, func(i, j int) bool {
			im, jm := s[i].rule.index, s[j].rule.index
			if im < jm {
				return true
			}
			if jm < im {
				return false
			}
			return s[i].position < s[j].position
		})
	}
	return &builder{
		state: flipped,
		seen:  p.toks,
	}
}

func (p *matcher) flipState() [][]item {
	flipped := make([][]item, len(p.state))
	for i, set := range p.state {
		for _, x := range set {
			if !x.complete() {
				continue
			}
			flipped[x.position] = append(flipped[x.position], item{
				rule:     x.rule,
				position: i,
				progress: x.progress,
			})
		}
	}
	return flipped
}

func (b *builder) build(root *symbol) (reflect.Value, error) {
	for _, top := range b.state[0] {
		if top.rule.implements != root {
			continue
		}
		if top.position != len(b.seen) {
			continue
		}
		span, ok := b.findSpan(top, 0)
		if !ok {
			return reflect.Value{}, ErrFailedMatch
		}
		return b.buildFromSpan(span)
	}
	return reflect.Value{}, ErrFailedMatch
}

func (b *builder) findSpan(x item, at int) (span, bool) {
	children, ok := b.findSpanChildren(x.rule.deps, at, x.position)
	if !ok {
		return span{}, false
	}
	return span{
		item:     x,
		at:       at,
		children: children,
	}, true
}

func (b *builder) buildFromSpan(s span) (reflect.Value, error) {
	if s.value.IsValid() {
		return s.value, nil
	}
	r := s.item.rule
	args := make([]reflect.Value, len(s.children)+1)
	args[0] = r.host
	for i, c := range s.children {
		child, err := b.buildFromSpan(c)
		if err != nil {
			return reflect.Value{}, err
		}
		args[i+1] = child
	}

	rets := r.method(r.host, args)
	if len(rets) == 2 && !rets[1].IsNil() {
		return reflect.Value{}, rets[1].Interface().(error)
	}
	return rets[0], nil
}

func (b *builder) findSpanChildren(deps []*symbol, at, end int) ([]span, bool) {
	if len(deps) == 0 {
		return nil, at == end
	}
	if deps[0].tokenType != nil {
		return b.tokenSpan(deps, at, end)
	}
	return b.ruleSpan(deps, at, end)
}

func (b *builder) ruleSpan(deps []*symbol, at, end int) ([]span, bool) {
	sym := deps[0]
	for _, found := range b.state[at] {
		if found.rule.implements != sym {
			continue
		}
		next, ok := b.findSpanChildren(deps[1:], found.position, end)
		if !ok {
			continue
		}
		inner, ok := b.findSpan(found, at)
		if !ok {
			continue
		}
		return append([]span{inner}, next...), true
	}
	return nil, false
}

func (b *builder) tokenSpan(deps []*symbol, at, end int) ([]span, bool) {
	sym := deps[0]
	if at >= len(b.seen) {
		return nil, false
	}
	if b.seen[at].Type().AssignableTo(sym.tokenType) {
		next, ok := b.findSpanChildren(deps[1:], at+1, end)
		if ok {
			return append([]span{{
				value: b.seen[at],
				at:    at,
			}}, next...), true
		}
	}
	return nil, false
}
