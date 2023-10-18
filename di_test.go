package di_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/mcvoid/di"
)

// a test value which can be bound by Inject()
type testBinder struct {
	wasCalled bool
	t         *testing.T
}

func (b *testBinder) Bind(f *os.File) {
	b.wasCalled = true
	if f != os.Stdin {
		b.t.Errorf("expected %v got %v", os.Stdin, f)
	}
}

func TestAdd(t *testing.T) {
	t.Run("doesn't panic on nil input", func(t *testing.T) {
		defer func() {
			val := recover()
			if val != nil {
				t.Errorf("expected no panic got %v", val)
			}
		}()
		di.New().Add(nil)
	})

	t.Run("zero value doesn't panic", func(t *testing.T) {
		defer func() {
			val := recover()
			if val != nil {
				t.Errorf("expected no panic got %v", val)
			}
		}()
		var ctx di.Context
		ctx.Add(os.Stdin)
	})
}

func TestInject(t *testing.T) {
	t.Run("nil injectee", func(t *testing.T) {
		ctx := di.New().Add(os.Stdout)

		defer func() {
			val := recover()
			if val != nil {
				t.Errorf("expected no panic got %v", val)
			}
		}()
		err := ctx.Inject(nil)
		if err == nil {
			t.Errorf("expected err got %v", err)
		}
	})

	t.Run("function no match", func(t *testing.T) {
		ctx := di.New().Add(os.Stdout)

		wasCalled := false
		fn := func(f io.WriterTo) {
			wasCalled = true
			if f != nil {
				t.Errorf("expected %v got %v", nil, f)
			}
		}

		err := ctx.Inject(fn)
		if !wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("zero value doesn't panic", func(t *testing.T) {
		var ctx di.Context

		wasCalled := false
		fn := func(f io.WriterTo) {
			wasCalled = true
			if f != nil {
				t.Errorf("expected %v got %v", nil, f)
			}
		}

		err := ctx.Inject(fn)
		if !wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("function exact type match", func(t *testing.T) {
		ctx := di.New().Add(os.Stdin)

		wasCalled := false
		fn := func(f *os.File) {
			wasCalled = true
			if f != os.Stdin {
				t.Errorf("expected %v got %v", os.Stdin, f)
			}
		}
		err := ctx.Inject(fn)
		if !wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("function interface implementation match", func(t *testing.T) {
		ctx := di.New().Add(os.Stdin)

		wasCalled := false
		fn := func(f io.Reader) {
			wasCalled = true
			if f != os.Stdin {
				t.Errorf("expected %v got %v", os.Stdin, f)
			}
		}

		err := ctx.Inject(fn)
		if !wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("function multiple exact match", func(t *testing.T) {
		var b bytes.Buffer
		ctx := di.New().Add(os.Stdin).Add(&b)

		wasCalled := false
		fn := func(f *os.File, buf *bytes.Buffer) {
			wasCalled = true
			if f != os.Stdin {
				t.Errorf("expected %v got %v", os.Stdin, f)
			}
			if buf != &b {
				t.Errorf("expected %v got %v", b, buf)
			}
		}

		err := ctx.Inject(fn)
		if !wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("function ambiguous match", func(t *testing.T) {
		var b bytes.Buffer
		ctx := di.New().Add(os.Stdout).Add(&b)

		fn := func(f io.Writer) {
			t.Errorf("expected func to not be called")
		}

		err := ctx.Inject(fn)
		if err == nil {
			t.Errorf("expected err got %v", err)
		}
	})

	t.Run("method exact match", func(t *testing.T) {
		ctx := di.New().Add(os.Stdin)

		b := testBinder{
			wasCalled: false,
			t:         t,
		}
		err := ctx.Inject(&b)
		if !b.wasCalled {
			t.Errorf("expected func to be called")
		}
		if err != nil {
			t.Errorf("expected %v got %v", nil, err)
		}
	})

	t.Run("function injecting into non-injectable", func(t *testing.T) {
		ctx := di.New().Add(os.Stdout)

		err := ctx.Inject(os.Stdout)
		if err == nil {
			t.Errorf("expected err got %v", err)
		}
	})
}
