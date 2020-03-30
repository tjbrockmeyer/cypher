package neohttp

type stats struct {
	ContainsUpdates_ bool `json:"contains_updates"`
	NodesCreated_ int `json:"nodes_created"`

	NodesDeleted_          int  `json:"nodes_deleted"`
	PropertiesSet_         int  `json:"properties_set"`
	RelationshipsCreated_  int  `json:"relationships_created"`
	RelationshipDeleted_   int  `json:"relationship_deleted"`
	LabelsAdded_           int  `json:"labels_added"`
	LabelsRemoved_         int  `json:"labels_removed"`
	IndexesAdded_          int  `json:"indexes_added"`
	IndexesRemoved_        int  `json:"indexes_removed"`
	ConstraintsAdded_      int  `json:"constraints_added"`
	ConstraintsRemoved_    int  `json:"constraints_removed"`
	ContainsSystemUpdates_ bool `json:"contains_system_updates"`
	SystemUpdates_         int  `json:"system_updates"`
}

func (s *stats) ConstraintsAdded() int {
	return s.ConstraintsAdded_
}

func (s *stats) ConstraintsRemoved() int {
	return s.ConstraintsRemoved_
}

func (s *stats) ContainsUpdates() bool {
	return s.ContainsUpdates_
}

func (s *stats) IndexesAdded() int {
	return s.IndexesAdded_
}

func (s *stats) IndexesRemoved() int {
	return s.IndexesRemoved_
}

func (s *stats) LabelsAdded() int {
	return s.LabelsAdded_
}

func (s *stats) LabelsRemoved() int {
	return s.LabelsRemoved_
}

func (s *stats) NodesCreated() int {
	return s.NodesCreated_
}

func (s *stats) NodesDeleted() int {
	return s.NodesDeleted_
}

func (s *stats) PropertiesSet() int {
	return s.PropertiesSet_
}

func (s *stats) RelationshipDeleted() int {
	return s.RelationshipDeleted_
}

func (s *stats) RelationshipsCreated() int {
	return s.RelationshipsCreated_
}

func (s *stats) ContainsSystemUpdates() bool {
	return s.ContainsSystemUpdates_
}

func (s *stats) SystemUpdates() int {
	return s.SystemUpdates_
}

