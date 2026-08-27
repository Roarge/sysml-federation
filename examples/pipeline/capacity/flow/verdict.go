package flow

import (
	"errors"
	"strconv"
	"strings"
)

// Names are the only two words of the model the service knows.
type Names struct{ Quantity, Attribute string }

// Subject is a requirement's subject as the router delivers it.
type Subject struct {
	Name         string
	HasAttribute bool
	Attribute    *float64
	Children     []Node
	Edges        []Edge
}

// The four verdict kinds, mirroring VerdictKind of the SysML v2 library.
const (
	KindPass         = "PASS"
	KindFail         = "FAIL"
	KindInconclusive = "INCONCLUSIVE"
	KindError        = "ERROR"
)

// Verdict applies the precedence of the capacity model: another quantity or no
// limit is INCONCLUSIVE before anything is looked at, then ERROR for a bad
// child, then the remaining INCONCLUSIVE cases, then PASS or FAIL. Every reason
// is one of the model's templates.
func Verdict(names Names, quantity, comparison string, limit *float64, subject Subject, verificationCase string) (kind, reason string) {
	if quantity != names.Quantity || limit == nil {
		if verificationCase != "" {
			return KindInconclusive, verificationCase + " is declared and no service runs it"
		}
		return KindInconclusive, "no service computes " + quantity
	}
	if len(subject.Children) > 0 {
		res, err := Rollup(subject.Children, subject.Edges)
		var ve *ValueError
		switch {
		case errors.As(err, &ve):
			return KindError, ve.Name + " has " + fault(ve) + " " + names.Attribute
		case errors.Is(err, ErrNoEntry):
			return KindInconclusive, "no entry part"
		case errors.Is(err, ErrNoExit):
			return KindInconclusive, "no exit part"
		}
		return judge(res.Capacity, comparison, *limit),
			names.Quantity + " " + num(res.Capacity) + " against " + num(*limit) + ", limited by " + strings.Join(cutNames(subject.Children, res.Cut), ", ")
	}
	if !subject.HasAttribute {
		return KindInconclusive, "no children to analyse"
	}
	if subject.Attribute == nil {
		return KindError, subject.Name + " has missing " + names.Attribute
	}
	if *subject.Attribute < 0 {
		return KindError, subject.Name + " has negative " + names.Attribute
	}
	return judge(*subject.Attribute, comparison, *limit), names.Attribute + " " + num(*subject.Attribute) + " against " + num(*limit)
}

// Analyse is what Part.capacity and Part.bottleneck read: the number and the
// cut where the wiring gives them, nil and empty otherwise.
func Analyse(subject Subject) (*float64, []string) {
	if len(subject.Children) > 0 {
		if res, err := Rollup(subject.Children, subject.Edges); err == nil {
			return &res.Capacity, res.Cut
		}
		return nil, nil
	}
	if subject.HasAttribute && subject.Attribute != nil && *subject.Attribute >= 0 {
		return subject.Attribute, nil
	}
	return nil, nil
}

func judge(value float64, comparison string, limit float64) string {
	var ok bool
	switch comparison {
	case "GE":
		ok = value >= limit
	case "GT":
		ok = value > limit
	case "LE":
		ok = value <= limit
	case "LT":
		ok = value < limit
	case "EQ":
		ok = value == limit
	}
	if ok {
		return KindPass
	}
	return KindFail
}

func fault(e *ValueError) string {
	if e.Negative {
		return "negative"
	}
	return "missing"
}

func num(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

// cutNames renders the cut in the order the router delivered the children.
func cutNames(children []Node, cut []string) []string {
	names := make([]string, 0, len(cut))
	for _, id := range cut {
		for _, c := range children {
			if c.ID == id {
				names = append(names, c.Name)
			}
		}
	}
	return names
}
