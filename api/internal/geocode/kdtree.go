package geocode

import "sort"

// KDTree is a 2D k-d tree over (lat, lon) points.
type KDTree struct {
	root *kdNode
}

type kdNode struct {
	city        *City
	axis        int // 0 = lat, 1 = lon
	left, right *kdNode
}

// BuildKDTree constructs a balanced k-d tree. The input slice is not modified.
func BuildKDTree(cities []City) *KDTree {
	if len(cities) == 0 {
		return &KDTree{}
	}
	ptrs := make([]*City, len(cities))
	for i := range cities {
		ptrs[i] = &cities[i]
	}
	return &KDTree{root: buildKD(ptrs, 0)}
}

func buildKD(pts []*City, depth int) *kdNode {
	if len(pts) == 0 {
		return nil
	}
	axis := depth % 2
	sort.Slice(pts, func(i, j int) bool {
		if axis == 0 {
			return pts[i].Lat < pts[j].Lat
		}
		return pts[i].Lon < pts[j].Lon
	})
	mid := len(pts) / 2
	return &kdNode{
		city:  pts[mid],
		axis:  axis,
		left:  buildKD(pts[:mid], depth+1),
		right: buildKD(pts[mid+1:], depth+1),
	}
}

// Nearest returns the city with smallest squared (lat,lon) distance to (lat, lon).
// Returns ok=false if the tree is empty.
func (t *KDTree) Nearest(lat, lon float64) (*City, bool) {
	if t == nil || t.root == nil {
		return nil, false
	}
	var best *City
	bestD := -1.0
	t.root.nearest(lat, lon, &best, &bestD)
	return best, best != nil
}

func (n *kdNode) nearest(lat, lon float64, best **City, bestD *float64) {
	if n == nil {
		return
	}
	dlat := lat - float64(n.city.Lat)
	dlon := lon - float64(n.city.Lon)
	d := dlat*dlat + dlon*dlon
	if *bestD < 0 || d < *bestD {
		*bestD = d
		*best = n.city
	}

	var diff float64
	if n.axis == 0 {
		diff = dlat
	} else {
		diff = dlon
	}

	near, far := n.left, n.right
	if diff > 0 {
		near, far = n.right, n.left
	}
	near.nearest(lat, lon, best, bestD)
	if diff*diff < *bestD {
		far.nearest(lat, lon, best, bestD)
	}
}
