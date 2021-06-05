package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/term"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

func initUserFlags() []cli.Flag {
	return append([]cli.Flag{
		&cli.StringFlag{
			Name:     "phone",
			Required: true,
			Usage:    "user phone",
			EnvVars:  []string{"PHONE"},
		},
	}, initFlags()...)
}

// terminalAuth implements auth.UserAuthenticator prompting the terminal for
// input.
type terminalAuth struct {
	phone string
}

func (terminalAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, xerrors.New("not implemented")
}

func (terminalAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (t terminalAuth) Phone(_ context.Context) (string, error) {
	return t.phone, nil
}

func (terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}

func (terminalAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	for {
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.TrimSpace(code)

		type notFlashing interface {
			GetLength() int
		}

		codeType, ok := sentCode.Type.(notFlashing)
		if ok && codeType.GetLength() != len(code) {
			fmt.Println("Code length must be", codeType.GetLength())
			continue
		}
		return code, nil
	}
}

func initUserCmd(c *cli.Context) error {
	phone := c.String("phone")
	if phone == "" {
		return xerrors.Errorf("no phone provided")
	}

	sessionName := fmt.Sprintf("gotd.session.%x.json", md5.Sum([]byte(phone))) // #nosec
	return genericInit(c, sessionName, func(ctx context.Context, client *telegram.Client) error {
		flow := auth.NewFlow(terminalAuth{phone: phone}, auth.SendCodeOptions{})

		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return xerrors.Errorf("auth: %w", err)
		}

		s, err := client.Auth().Status(ctx)
		if err != nil {
			return xerrors.Errorf("status: %w", err)
		}

		if s.User != nil {
			as := strconv.Itoa(s.User.ID)
			switch {
			case s.User.Username != "":
				as = "@" + s.User.Username
			case s.User.FirstName != "":
				as = s.User.FirstName + " " + s.User.LastName
			}

			fmt.Printf("Logged as user %s\n", as)
		}
		return nil
	})
}
