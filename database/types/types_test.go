package types

import (
	"bytes"
	"testing"
)

var (
	results = []*ServerSession{
		&ServerSession{
			[]byte("6e2963fa0c3b2d2baaa695f963a9081e91f26e7bd25623989655ada3ab3e76a5"),
			[]byte("TuTRM6NreQb0UJ442Z2TaUFVqfhTWX/e45WqlVCYu+w="),
			[]byte("eric@sermodigital.com"),
			1433284583204993508,
		},
		&ServerSession{
			[]byte("6ade419d808796c7a5ad14810c97f75153f4fa78bf8a7bf15e3382ae6dd583de"),
			[]byte("XbcXGHC4/4RKiaAHMWz7IyLwxtQaAzA1N/TzhPuGdJ0="),
			[]byte("eric@sermodigital.com"),
			1433284583205005041,
		},
		&ServerSession{
			[]byte("bd38767c324e95295376571e398e29750a09a71fddbbbc27addf9cae22c6f9e4"),
			[]byte("G14NCrRh0KfNd82Tcv75gO9hkLP38CBvUZ0rFuK04ws="),
			[]byte("eric@sermodigital.com"),
			1433284583205015369,
		},
		&ServerSession{
			[]byte("7f0f98ec005618453a6c272400c21ff9a2a61ef7e78a3c4b66d5043f4678ad31"),
			[]byte("zideBSeTVwmMe8YOjG+KO04EyBfULdHBnRhtr6GoDHk="),
			[]byte("eric@sermodigital.com"),
			1433284583205025528,
		},
		&ServerSession{
			[]byte("549b10cd8958b97127dec02a625ca02648c0b86b733d96485ee408521ec9b896"),
			[]byte("/khUv/ZH6G5MSmrNFWIViati0JKrgIBy9jPoB/4WWSw="),
			[]byte("eric@sermodigital.com"),
			1433284583205035635,
		},
		&ServerSession{
			[]byte("17d6e0445baa84d9cc5f5a71421b6372f9e5bc73fdff08f7a21ae83ae4efecc0"),
			[]byte("3hg8VV4g97nH/qi3NIuHvM2upMWEb3UwE26f0prfVJk="),
			[]byte("eric@sermodigital.com"),
			1433284583205045706,
		},
		&ServerSession{
			[]byte("b026116f4707c2c01d0316d58ba20d68ac4a6dd3b0297bb47e9cc0370a0c236d"),
			[]byte("I29+0JD91viI/dnOdhnJvM69eUJAzFaopOxpwnj0cCk="),
			[]byte("eric@sermodigital.com"),
			1433284583205055765,
		},
		&ServerSession{
			[]byte("a8b804a237340488f638e1f0be1f2b743bcc1606490eebb312b7b5dcb9263770"),
			[]byte("LuN27sjjrJE9AX/lCrXucbO5389u6jHWeQIFgvCxitE="),
			[]byte("eric@sermodigital.com"),
			1433284583205065803,
		},
		&ServerSession{
			[]byte("bcd0446fb1f20e337986861b2f482058888f5d25c22169c4c497c66d9d3b0400"),
			[]byte("ltBzQErCgQTDn3qrXHy1AS2+gUqqGwv+PbA3m6iq5to="),
			[]byte("eric@sermodigital.com"),
			1433284583205075860,
		},
		&ServerSession{
			[]byte("4fe483e8bad6fb604ede1845cfafdb9d5a5d8d70d86e391f13b37270473ad8bb"),
			[]byte("AbSCe0u0AeTDytCFdD/33KLhZMGa9gM+OCUt6oVeH6Y="),
			[]byte("eric@sermodigital.com"),
			1433284583205085894,
		},
	}

	temp = make([]ServerSession, len(results))
	ms   = make([][]byte, len(results))
)

func TestMarshal(t *testing.T) {
	for i, v := range results {
		res, err := v.MarshalJSON()
		if err != nil {
			t.Error(err)
		}

		ms[i] = res
	}
}

func TestUnmarshal(t *testing.T) {
	for i, v := range ms {
		err := temp[i].UnmarshalJSON(v)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestMarshalEqual(t *testing.T) {
	for i, v := range temp {
		if !bytes.Equal(temp[i].AuthToken, v.AuthToken) {
			t.Error("Unmarshal AuthToken isn't equal")
		}
		if !bytes.Equal(temp[i].CSRFToken, v.CSRFToken) {
			t.Error("Unmarshal CSRFToken isn't equal")
		}
		if !bytes.Equal(temp[i].Email, v.Email) {
			t.Error("Unmarshal Email isn't equal")
		}
		if temp[i].Date != v.Date {
			t.Error("Unmarshal Date isn't equal")
		}
	}
}
