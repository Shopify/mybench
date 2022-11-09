package mybench

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPushToFullRingAndReadAll(t *testing.T) {
	ring := NewRing[int](10)
	for i := 1; i < 12; i++ {
		ring.Push(i)
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, data)

	ring.Push(12)
	data = ring.ReadAllOrdered()
	require.Equal(t, []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, data)
}

func TestPushPartialRingAndReadAll(t *testing.T) {
	ring := NewRing[int](10)
	for i := 1; i < 5; i++ {
		ring.Push(i)
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, []int{1, 2, 3, 4}, data)
}

func TestReadAllFromEmptyRing(t *testing.T) {
	ring := NewRing[int](10)
	data := ring.ReadAllOrdered()
	require.Equal(t, []int{}, data)
}

type A struct {
	Data1 int
	Data2 int
	Data3 int
}

func TestReadAllFromPartialRingWithStruct(t *testing.T) {
	ring := NewRing[A](10)
	for i := 1; i < 3; i++ {
		ring.Push(A{Data1: i, Data2: i * 2, Data3: i * 3})
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, []A{
		{Data1: 1, Data2: 2, Data3: 3},
		{Data1: 2, Data2: 4, Data3: 6},
	}, data)
}

func TestReadAllFromFullRingWithStruct(t *testing.T) {
	ring := NewRing[A](4)
	for i := 1; i < 6; i++ {
		ring.Push(A{Data1: i, Data2: i * 2, Data3: i * 3})
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, []A{
		{Data1: 2, Data2: 4, Data3: 6},
		{Data1: 3, Data2: 6, Data3: 9},
		{Data1: 4, Data2: 8, Data3: 12},
		{Data1: 5, Data2: 10, Data3: 15},
	}, data)
}

func TestReadAllFromPartialRingWithStructPointers(t *testing.T) {
	ring := NewRing[*A](10)
	for i := 1; i < 3; i++ {
		ring.Push(&A{Data1: i, Data2: i * 2, Data3: i * 3})
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, 2, len(data))
	require.Equal(t, []*A{
		{Data1: 1, Data2: 2, Data3: 3},
		{Data1: 2, Data2: 4, Data3: 6},
	}, data)
}

func TestReadAllFromFullRingWithStructPointers(t *testing.T) {
	ring := NewRing[*A](4)
	for i := 1; i < 6; i++ {
		ring.Push(&A{Data1: i, Data2: i * 2, Data3: i * 3})
	}

	data := ring.ReadAllOrdered()
	require.Equal(t, []*A{
		{Data1: 2, Data2: 4, Data3: 6},
		{Data1: 3, Data2: 6, Data3: 9},
		{Data1: 4, Data2: 8, Data3: 12},
		{Data1: 5, Data2: 10, Data3: 15},
	}, data)
}
