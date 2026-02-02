package processors

type AuthType string

const (
	AuthTypeServiceAccount AuthType = "serviceAccount"
)

type K8sAttributesProcessor struct {
	typeName string

	AuthType AuthType             `json:"authType,omitempty"`
	Filter   *Filter              `json:"filter,omitempty"`
	Extract  *Extract             `json:"extract,omitempty"`
	Labels   []AttributeExtractor `json:"labels,omitempty"`
}

type Filter struct {
	NodeFromEnvVar string `json:"node_from_env_var,omitempty"`
}

type Extract struct {
	Metadata []string `json:"metadata,omitempty"`
}
type AttibuteSource string

const (
	AttributeSourceDeployment AttibuteSource = "deployment"
	AttributeSourceNamespace  AttibuteSource = "namespace"
	AttributeSourceNode       AttibuteSource = "node"
	AttributeSourcePod        AttibuteSource = "pod"
)

type AttributeExtractor struct {
	TagName  string `json:"tag_name,omitempty"`
	Key      string `json:"key,omitempty"`
	KeyRegex string `json:"key_regex,omitempty"`
	From     string `json:"from,omitempty"`
}
