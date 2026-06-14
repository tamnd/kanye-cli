package kanye

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes kanye as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/kanye-cli/kanye"
func init() { kit.Register(Domain{}) }

// Domain is the kanye driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "kanye",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "kanye",
			Short:  "Random Kanye West quotes from api.kanye.rest",
			Long: `kanye fetches random Kanye West quotes from the public api.kanye.rest API.
No login or API key required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/kanye-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// quote: fetch one random Kanye quote
	kit.Handle(app, kit.OpMeta{
		Name:    "quote",
		Group:   "read",
		Single:  true,
		Summary: "Fetch one random Kanye West quote",
	}, quoteOp)

	// quotes: fetch N random Kanye quotes
	kit.Handle(app, kit.OpMeta{
		Name:    "quotes",
		Group:   "read",
		List:    true,
		Summary: "Fetch multiple random Kanye West quotes",
	}, quotesOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type quoteInput struct {
	Client *Client `kit:"inject"`
}

type quotesInput struct {
	Count  int     `kit:"flag,inherit" help:"number of quotes to fetch (default 5)"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func quoteOp(ctx context.Context, in quoteInput, emit func(Quote) error) error {
	q, err := in.Client.Quote(ctx)
	if err != nil {
		return err
	}
	return emit(q)
}

func quotesOp(ctx context.Context, in quotesInput, emit func(Quote) error) error {
	count := in.Count
	if count <= 0 {
		count = 5
	}
	items, err := in.Client.Quotes(ctx, count)
	if err != nil {
		return err
	}
	for _, q := range items {
		if err := emit(q); err != nil {
			return err
		}
	}
	return nil
}

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty kanye reference")
	}
	return "quote", input, nil
}

// Locate returns the live URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "quote":
		return "https://api.kanye.rest/", nil
	default:
		return "", errs.Usage("kanye has no resource type %q", uriType)
	}
}
