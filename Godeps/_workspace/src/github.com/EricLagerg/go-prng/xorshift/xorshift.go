/*
   Written in Go 2015 by Eric Lagergren (contact@ericlagergren.com)
   Written in C 2014 by Sebastiano Vigna (vigna@acm.org) and
   Kenji Rikitake (kenji.rikitake@acm.org).

To the extent possible under law, the author has dedicated all copyright
and related and neighboring rights to this software to the public domain
worldwide. This software is distributed without any warranty.

See <http://creativecommons.org/publicdomain/zero/1.0/>. */
package prng

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

// Guarenteed to be fair using a dice roll.
const randSeed = 16256612229375771919

type XORShift interface {
	Next() uint64
	Seed()
}

// randUint reads from the OS' crypto PRNG and uses its output
// to create a slice of uint64s with the given lenght, n.
func randUint(n int) []uint64 {
	s := make([]uint64, n)

	// Fill the array the user gave us but make sure there aren't any zeros.
	for i := 0; i < n; i++ {
		s[i] = randomNonZero()
	}

	return s
}

// randomNonZero fetches 8 bytes from the OS' CSPRNG and returns it as
// a non-zero uint64. Will panic if it can't read from rand.Reader.
func randomNonZero() uint64 {
	buf := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != 8 {
		panic("Unable to fully read from rand.Reader")
	}
	u, x := binary.Uvarint(buf)
	if u == 0 || x == 0 || x < 0 {
		return randomNonZero()
	}
	return u
}

// fillWithXOR fills a slice using a xorshift64 generator, using a
// determined starting state (a large prime).
func fillWithXOR(n int) []uint64 {
	s := make([]uint64, n)
	r := new(Shift64Star)
	r.x = randSeed
	for i := range s {
		s[i] = r.Next()
	}
	return s
}

// genState generates a starting state using Go's 'crypto/rand'
// package. The state will be a 64-bit prime. It'll panic if
// rand.Prime returns an error.
func genState() uint64 {
	prime, err := rand.Prime(rand.Reader, 64)
	if err != nil {
		panic(err)
	}
	return prime.Uint64()
}

const uint58mask uint64 = (1 << 58) - 1

// Shift116Plus is a variant of xorshift128+ for dynamic languages, such
// as Erlang, that can use only 58 bits of a 64-bit integer. Only the lower
// 58 bits of each state word are valid (the upper six are zeroes).
//
// This generator passes BigCrush without systematic failures, but due to
// the relatively short period it is acceptable only for applications with
// a mild amount of parallelism; otherwise, use a xorshift1024* generator.
//
// The state must be seeded so that the lower 58 bits of s[ 0 ] and s[ 1 ]
// are not all zeroes. If you have a nonzero 64-bit seed, we suggest to
// pass it twice through MurmurHash3's avalanching function and take the
// lower 58 bits, taking care that they are not all zeroes (you can apply
// the avalanching function again if this happens).
type Shift116Plus struct {
	state [2]uint64
}

func (s *Shift116Plus) Next() uint64 {
	s0 := s.state[1]
	s1 := s.state[0]
	s.state[0] = s0
	s1 ^= (s1 << 24) & uint58mask // a
	s.state[1] = (s1 ^ s0 ^ (s1 >> 11) ^ (s0 >> 41))
	return (((s.state[1]) + s0) & uint58mask) // b, c
}

func (s *Shift116Plus) Seed() {
	copy(s.state[:], randUint(cap(s.state)))
}

// Shift128Plus is the fastest generator passing BigCrush without
// systematic failures, but due to the relatively short period it is
// acceptable only for applications with a mild amount of parallelism;
// otherwise, use a xorshift1024* generator.
//
// The state must be seeded so that it is not everywhere zero. If you have
// a nonzero 64-bit seed, we suggest to pass it twice through
// MurmurHash3's avalanching function.
type Shift128Plus struct {
	state [2]uint64
}

func (s *Shift128Plus) Next() uint64 {
	s0 := s.state[1]
	s1 := s.state[0]
	s.state[0] = s0
	s1 ^= s1 << 23 // a
	s.state[1] = (s1 ^ s0 ^ (s1 >> 17) ^ (s0 >> 26))
	return (s.state[1]) + s0 // b, c
}

func (s *Shift128Plus) Seed() {
	copy(s.state[:], randUint(cap(s.state)))
}

// Shift1024Star is a fast, top-quality generator. If 1024 bits of state are
// too much, try a xorshift128+ or generator.
//
// The state must be seeded so that it is not everywhere zero. If you have
// a 64-bit seed,  we suggest to seed a xorshift64* generator and use its
// output to fill s.
type Shift1024Star struct {
	state [16]uint64
	p     int
}

func (s *Shift1024Star) Next() uint64 {
	s0 := s.state[s.p]
	s.p = (s.p + 1) & 15
	s1 := s.state[s.p]
	s1 ^= s1 << 31
	s1 ^= s1 >> 11
	s0 ^= s0 >> 30
	s.state[s.p] = s0 ^ s1
	return s.state[s.p] * 1181783497276652981
}

func (s *Shift1024Star) Seed() {
	copy(s.state[:], fillWithXOR(cap(s.state)))
	s.p = 0
}

// Shift40956 is usable, but we suggest you use a
// xorshift1024* generator.
//
// The state must be seeded so that it is not everywhere zero. If you have
// a 64-bit seed,  we suggest to seed a xorshift64* generator and use its
// output to fill s.
type Shift4096Star struct {
	state [64]uint64
	p     int
}

func (s *Shift4096Star) Next() uint64 {
	s0 := s.state[s.p]
	s.p = (s.p + 1) & 63
	s1 := s.state[s.p]
	s1 ^= s1 << 25 // a
	s1 ^= s1 >> 3  // b
	s0 ^= s0 >> 49 // c
	s.state[s.p] = s0 ^ s1
	return (s.state[s.p]) * 8372773778140471301
}

func (s *Shift4096Star) Seed() {
	copy(s.state[:], fillWithXOR(cap(s.state)))
	s.p = 0
}

// Shift64Star is a fast, good generator if you're short on memory, but
// otherwise we rather suggest to use a xorshift128+ or xorshift1024*
// (for a very long period) generator.
type Shift64Star struct {
	x uint64 // state
}

func (s *Shift64Star) Next() uint64 {
	s.x ^= s.x >> 12 // a
	s.x ^= s.x << 25 // b
	s.x ^= s.x >> 27 // c
	return s.x * 2685821657736338717
}

func (s *Shift64Star) Seed() {
	s.x = genState()
}
