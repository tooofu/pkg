package xnebula

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

// nebula ResultSet -> OpenCypher style

const (
	entityTypeNode = "node"
	entityTypeEdge = "relationship"
)

//	type Result struct {
//	    Data []map[string]interface{} `json:"results,omitempty"`
//	}
// type Result map[string]interface{}

type Node struct {
	ID         string            `json:"~id,omitempty"`
	EntityType string            `json:"~entityType,omitempty"`
	Labels     []string          `json:"~labels,omitempty"`
	Properties map[string]string `json:"~properties,omitempty"`
}

type Edge struct {
	ID         string            `json:"~id,omitempty"`
	EntityType string            `json:"~entityType,omitempty"`
	Start      string            `json:"~start,omitempty"`
	End        string            `json:"~end,omitempty"`
	Type       string            `json:"~type,omitempty"`
	Properties map[string]string `json:"~properties,omitempty"`
}

type Path []interface{}

func ConvertResult(nr *nebula.ResultSet) (rs []map[string]interface{}, err error) {
	cols := nr.GetColNames()
	rows := nr.GetRowSize()

	var (
		// record  *nebula.Record
		valwrap *nebula.ValueWrapper
		value   interface{}
	)

	for i := 0; i < rows; i++ {
		var record *nebula.Record
		if record, err = nr.GetRowValuesByIndex(i); err != nil {
			return
		}

		m := make(map[string]interface{})
		for j, col := range cols {
			if valwrap, err = record.GetValueByIndex(j); err != nil {
				return
			}
			if value, err = convValue(valwrap); err != nil {
				return
			}
			m[col] = value
		}
		rs = append(rs, m)
	}

	return
}

func convValue(v *nebula.ValueWrapper) (r interface{}, err error) {
	var (
		node *nebula.Node
		edge *nebula.Relationship
		path *nebula.PathWrapper
	)

	switch {
	default:
		r = v.String()
	case v.IsVertex():
		if node, err = v.AsNode(); err != nil {
			return
		}
		r, err = convNode(node)
	case v.IsEdge():
		if edge, err = v.AsRelationship(); err != nil {
			return
		}
		r, err = convEdge(edge)
	case v.IsPath():
		if path, err = v.AsPath(); err != nil {
			return
		}
		r, err = convPath(path)
	case v.IsList():
		var (
			rs  = make([]interface{}, 0)
			vws []nebula.ValueWrapper
			val interface{}
		)
		if vws, err = v.AsList(); err != nil {
			return
		}
		for _, vw := range vws {
			if val, err = convValue(&vw); err != nil {
				return
			}
			rs = append(rs, val)
		}
		r = rs
	case v.IsMap():
		var (
			rp  = make(map[string]interface{})
			vwp map[string]nebula.ValueWrapper
			val interface{}
		)
		if vwp, err = v.AsMap(); err != nil {
			return
		}
		for key, vw := range vwp {
			if val, err = convValue(&vw); err != nil {
				return
			}
			rp[key] = val
		}
		r = rp
	case v.IsInt():
		r, err = v.AsInt()
	}
	return
}

func convNode(n *nebula.Node) (r *Node, err error) {
	r = &Node{
		ID:         n.GetID().String(),
		EntityType: entityTypeNode,
		Labels:     n.GetTags(),
		Properties: make(map[string]string),
	}

	for _, t := range r.Labels {
		var vwp map[string]*nebula.ValueWrapper

		if vwp, err = n.Properties(t); err != nil {
			return
		}

		for key, val := range vwp {
			r.Properties[key] = val.String()
		}

		// first tag only
		break
	}

	return
}

func convEdge(e *nebula.Relationship) (r *Edge, err error) {
	r = &Edge{
		ID:         fmt.Sprintf("%s -> %s", e.GetSrcVertexID().String(), e.GetDstVertexID().String()),
		EntityType: entityTypeEdge,
		Start:      e.GetSrcVertexID().String(),
		End:        e.GetDstVertexID().String(),
		Properties: make(map[string]string),
	}

	for key, val := range e.Properties() {
		r.Properties[key] = val.String()
	}

	return
}

func convPath(p *nebula.PathWrapper) (r *Path, err error) {
	ps := make(Path, 0)
	N := len(p.GetNodes()) + len(p.GetRelationships())
	var (
		left, right int

		nodes []*nebula.Node
		edges []*nebula.Relationship
		n     *Node
		e     *Edge
	)
	nodes = p.GetNodes()
	edges = p.GetRelationships()

	for i := 0; i < N; i++ {
		if i%2 == 0 {
			if n, err = convNode(nodes[left]); err != nil {
				return
			}
			ps = append(ps, n)
			left++
		} else {
			if e, err = convEdge(edges[right]); err != nil {
				return
			}
			ps = append(ps, e)
			right++
		}
	}

	return
}
