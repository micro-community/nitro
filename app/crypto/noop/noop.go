package noop

import (
	"github.com/gonitro/nitro/app/crypto"
	"github.com/gonitro/nitro/util/uuid"
)

func NewAuth(opts ...crypto.Option) crypto.Auth {
	var options crypto.Options
	for _, o := range opts {
		o(&options)
	}

	return &noop{
		opts: options,
	}
}

func NewRules() crypto.Rules {
	return &noopRules{}
}

type noop struct {
	opts crypto.Options
}

// String returns the name of the implementation
func (n *noop) String() string {
	return "noop"
}

// Init the crypto
func (n *noop) Init(opts ...crypto.Option) {
	for _, o := range opts {
		o(&n.opts)
	}
}

// Options set for crypto
func (n *noop) Options() crypto.Options {
	return n.opts
}

// Generate a new account
func (n *noop) Generate(id string, opts ...crypto.GenerateOption) (*crypto.Account, error) {
	options := crypto.NewGenerateOptions(opts...)
	name := options.Name
	if name == "" {
		name = id
	}
	return &crypto.Account{
		ID:       id,
		Secret:   options.Secret,
		Metadata: options.Metadata,
		Scopes:   options.Scopes,
		Issuer:   n.Options().Issuer,
		Name:     name,
	}, nil
}

// Inspect a token
func (n *noop) Inspect(token string) (*crypto.Account, error) {
	return &crypto.Account{ID: uuid.New().String(), Issuer: n.Options().Issuer}, nil
}

// Token generation using an account id and secret
func (n *noop) Token(opts ...crypto.TokenOption) (*crypto.Token, error) {
	return &crypto.Token{}, nil
}

type noopRules struct{}

// Grant access to a resource
func (n *noopRules) Grant(rule *crypto.Rule) error {
	return nil
}

// Revoke access to a resource
func (n *noopRules) Revoke(rule *crypto.Rule) error {
	return nil
}

func (n *noopRules) List(opts ...crypto.RulesOption) ([]*crypto.Rule, error) {
	return []*crypto.Rule{}, nil
}

// Verify an account has access to a resource
func (n *noopRules) Verify(acc *crypto.Account, res *crypto.Resource, opts ...crypto.VerifyOption) error {
	return nil
}
