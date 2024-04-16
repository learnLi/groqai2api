package accountpool

import (
	"errors"
	groq "github.com/learnLi/groq_client"
	"sync"
)

type IAccounts struct {
	Accounts []*groq.Account `json:"accounts"`
	mx       sync.Mutex
}

func (a *IAccounts) Get() *groq.Account {
	a.mx.Lock()
	defer a.mx.Unlock()
	if len(a.Accounts) == 0 {
		return nil
	}
	account := a.Accounts[0]
	a.Accounts = append(a.Accounts[1:], account)
	return account
}

func (a *IAccounts) GetList() []*groq.Account {
	return a.Accounts
}

func (a *IAccounts) Update(account *groq.Account) {
	a.mx.Lock()
	defer a.mx.Unlock()
	for i, v := range a.Accounts {
		if v.SessionToken == account.SessionToken {
			a.Accounts[i] = account
			return
		}
	}
}

func (a *IAccounts) Add(tokens []string) error {
	a.mx.Lock()
	defer a.mx.Unlock()
	if len(tokens) == 0 {
		return errors.New("tokens is empty")
	}
	existingTokens := make(map[string]struct{})
	for _, acc := range a.Accounts {
		existingTokens[acc.SessionToken] = struct{}{}
	}
	for _, token := range tokens {
		if _, exists := existingTokens[token]; !exists {
			a.Accounts = append(a.Accounts, groq.NewAccount(token, ""))
			existingTokens[token] = struct{}{} // Add to set to prevent duplicates within the input tokens.
		}
	}
	return nil
}

func NewAccounts(accounts []*groq.Account) *IAccounts {
	return &IAccounts{Accounts: accounts}
}
