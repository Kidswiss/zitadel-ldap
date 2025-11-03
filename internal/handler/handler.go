package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strings"

	"github.com/GeertJohan/yubigo"
	"github.com/glauth/glauth/v2/pkg/config"
	"github.com/glauth/glauth/v2/pkg/handler"
	"github.com/glauth/ldap"
	"github.com/rs/zerolog"

	"github.com/jacobalberty/zitadel-ldap/internal/zitadel"
)

type zitadelHandler struct {
	options   handler.Options
	ldohelper handler.LDAPOpsHelper
	// tracer        trace.Tracer
	zitadelClient *zitadel.Client
	log           *zerolog.Logger
	yubikeyAuth   *yubigo.YubiAuth
	baseDN        string
}

// NewZitadelHandler creates a new instance of the zitadel backend
func NewZitadelHandler(opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &zitadelHandler{
		options:       options,
		ldohelper:     options.LDAPHelper,
		log:           options.Logger,
		yubikeyAuth:   options.YubiAuth,
		zitadelClient: zitadel.NewClient(os.Getenv("ZITADEL_URL"), os.Getenv("ZITADEL_PAT"), nil, options.Logger),
		baseDN:        options.Backend.BaseDN,
	}
}

func (z *zitadelHandler) Add(boundDN string, req ldap.AddRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultInsufficientAccessRights, nil
}

func (z *zitadelHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	// _, span := z.tracer.Start(context.Background(), "plugins.zitadelHandler.Bind")
	// defer span.End()

	z.log.Info().Str("username", bindDN).Msg("trying to login")

	userCN := strings.Split(bindDN, ",")[0]
	user := strings.Split(userCN, "=")[1]

	success, err := z.zitadelClient.Login(user, bindSimplePw)
	if err != nil {
		return ldap.LDAPResultInvalidCredentials, err
	}

	if success {
		return ldap.LDAPResultSuccess, nil
	}

	return ldap.LDAPResultInvalidCredentials, nil
}

func (z *zitadelHandler) Search(boundDN string, req ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	// ctx, span := z.tracer.Start(context.Background(), "plugins.zitadelHandler.Search")
	// defer span.End()
	z.log.Info().Str("backend", "zitadel").Str("user", boundDN).Msg("Doing a search")
	return z.ldohelper.Search(context.Background(), z, boundDN, req, conn)
}

func (z *zitadelHandler) Close(boundDN string, conn net.Conn) error {
	return nil
}

func (z *zitadelHandler) Modify(boundDN string, req ldap.ModifyRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultInsufficientAccessRights, nil
}

func (z *zitadelHandler) Delete(boundDN, deleteDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultInsufficientAccessRights, nil
}

func (z *zitadelHandler) FindUser(ctx context.Context, userName string, searchByUPN bool) (bool, config.User, error) {

	z.log.Info().Str("backend", "zitadel").Str("username", userName).Msg("Finding user")

	var users *zitadel.UserResults
	if searchByUPN {
		tmp, err := z.zitadelClient.FindUserByMail(userName)
		if err != nil {
			return false, config.User{}, err
		}
		users = tmp
	} else {
		tmp, err := z.zitadelClient.FindUserByName(userName)
		if err != nil {
			return false, config.User{}, err
		}
		users = tmp
	}

	if users == nil || len(users.Result) == 0 || len(users.Result) > 1 {
		return false, config.User{}, nil
	}

	zUser := users.Result[0]

	metadata, err := z.zitadelClient.ListMetadata(zUser.UserID)
	if err != nil {
		return false, config.User{}, err
	}

	caps := []config.Capability{}

	z.log.Debug().Str("user", zUser.Username).Int("metadata", len(metadata.Result)).Msg("Found metadata")

	for _, m := range metadata.Result {
		z.log.Debug().Str("user", zUser.Username).Str("metadata key", m.Key).Str("value", string(m.Value)).Msg("Checking metadata key")
		if strings.HasPrefix(m.Key, "cap_") {
			z.log.Debug().Str("user", zUser.Username).Str("metadata key", m.Key).Msg("Valid cap key")
			key := strings.TrimPrefix(m.Key, "cap_")
			val, err := base64.StdEncoding.DecodeString(string(m.Value))
			if err != nil {
				return false, config.User{}, err
			}

			z.log.Debug().Str("user", zUser.Username).Str("metadata key", m.Key).Str("value", string(val)).Msg("Appending to existing caps")

			caps = append(caps, config.Capability{
				Action: key,
				Object: string(val),
			})
		}
	}

	z.log.Debug().Str("user", zUser.Username).Int("caps", len(caps)).Msg("Found capabilities")

	groups, err := z.getUserGroupIDs(zUser.UserID)
	if err != nil {
		return false, config.User{}, err
	}

	user := config.User{
		Name:         zUser.Username,
		Disabled:     zUser.State != "USER_STATE_ACTIVE",
		Mail:         zUser.Human.Email.Email,
		UnixID:       z.hash(zUser.Username),
		UIDNumber:    z.hash(zUser.Username),
		Capabilities: caps,
		OtherGroups:  groups,
	}

	return true, user, nil
}

func (z *zitadelHandler) FindGroup(ctx context.Context, groupName string) (bool, config.Group, error) {
	z.log.Info().Str("backend", "zitadel").Str("group", groupName).Msg("Finding groups")

	projects, err := z.zitadelClient.ListProjects()
	if err != nil {
		return false, config.Group{}, err
	}

	for _, p := range projects.Result {
		roles, err := z.zitadelClient.ListRoles(p.ID)
		if err != nil {
			return false, config.Group{}, err
		}

		for _, r := range roles.Result {
			if r.Group == groupName {
				return true, config.Group{
					Name:      groupName,
					UnixID:    z.hash(p.ID + groupName),
					GIDNumber: z.hash(p.ID + groupName),
				}, nil
			}
		}
	}

	return false, config.Group{}, nil
}

func (z *zitadelHandler) GetBackend() config.Backend       { return z.options.Backend }
func (z *zitadelHandler) GetLog() *zerolog.Logger          { return z.log }
func (z *zitadelHandler) GetCfg() *config.Config           { return z.options.Config }
func (z *zitadelHandler) GetYubikeyAuth() *yubigo.YubiAuth { return z.yubikeyAuth }

func (z *zitadelHandler) FindPosixAccounts(ctx context.Context, hierarchy string) (entrylist []*ldap.Entry, err error) {
	// _, span := z.tracer.Start(ctx, "handler.zitadelHandler.FindPosixAccounts")
	// defer span.End()
	z.log.Info().Str("backend", "zitadel").Str("posixAccount", hierarchy).Msg("Finding posixAccount")

	entries := []*ldap.Entry{}

	users, err := z.zitadelClient.ListUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users.Result {

		if u.Human == nil {
			continue
		}

		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{Name: "cn", Values: []string{u.Username}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "givenName", Values: []string{u.Human.Profile.GivenName}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "sn", Values: []string{u.Human.Profile.FamilyName}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "ou", Values: []string{"users"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "uidNumber", Values: []string{fmt.Sprintf("%d", z.hash(u.Username))}})
		if u.State != "USER_STATE_ACTIVE" {
			attrs = append(attrs, &ldap.EntryAttribute{Name: "accountStatus", Values: []string{"inactive"}})
		} else {
			attrs = append(attrs, &ldap.EntryAttribute{Name: "accountStatus", Values: []string{"active"}})
		}
		attrs = append(attrs, &ldap.EntryAttribute{Name: "mail", Values: []string{u.Human.Email.Email}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "userPrincipalName", Values: []string{u.Human.Email.Email}})

		attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"posixAccount", "shadowAccount"}})

		attrs = append(attrs, &ldap.EntryAttribute{Name: "loginShell", Values: []string{"/bin/bash"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "homeDirectory", Values: []string{"/home/" + u.Username}})

		attrs = append(attrs, &ldap.EntryAttribute{Name: "description", Values: []string{u.Username}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "gecos", Values: []string{u.Username}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "gidNumber", Values: []string{fmt.Sprintf("%d", z.hash("zitadel"))}})

		groups, err := z.memberOf(u.UserID)
		if err != nil {
			return nil, err
		}

		attrs = append(attrs, &ldap.EntryAttribute{Name: "memberOf", Values: groups})

		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowExpire", Values: []string{"-1"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowFlag", Values: []string{"134538308"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowInactive", Values: []string{"-1"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowLastChange", Values: []string{"11000"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowMax", Values: []string{"99999"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowMin", Values: []string{"-1"}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "shadowWarning", Values: []string{"7"}})

		metadata, err := z.zitadelClient.ListMetadata(u.UserID)
		if err != nil {
			return nil, err
		}

		for _, m := range metadata.Result {
			if strings.HasPrefix(m.Key, "gl_") {
				key := strings.TrimPrefix(m.Key, "gl_")
				val, err := base64.StdEncoding.DecodeString(string(m.Value))
				if err != nil {
					return nil, err
				}
				attrs = append(attrs, &ldap.EntryAttribute{Name: key, Values: []string{string(val)}})
			}
		}

		dn := fmt.Sprintf("%s=%s,%s,%s", z.options.Backend.NameFormat, u.Username, hierarchy, z.baseDN)

		entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
	}

	return entries, nil

}
func (z *zitadelHandler) FindPosixGroups(ctx context.Context, hierarchy string) (entrylist []*ldap.Entry, err error) {
	// _, span := z.tracer.Start(ctx, "handler.zitadelHandler.FindPosixGroups")
	// defer span.End()
	z.log.Info().Str("backend", "zitadel").Str("hieararchy", hierarchy).Msg("Finding posixGroup")

	asGroupOfUniqueNames := hierarchy == "ou=groups"

	projects, err := z.zitadelClient.ListProjects()
	if err != nil {
		return nil, err
	}

	grants, err := z.zitadelClient.ListGrants()
	if err != nil {
		return nil, err
	}

	grantMap := map[string][]string{}

	for _, g := range grants.Result {
		for _, r := range g.RoleKeys {
			grantMap[g.ProjectName+r] = append(grantMap[g.ProjectName+r], z.addUserHierarchy(g.UserName, hierarchy))
		}
	}

	entries := []*ldap.Entry{}

	for _, p := range projects.Result {
		roles, err := z.zitadelClient.ListRoles(p.ID)
		if err != nil {
			return nil, err
		}

		for _, r := range roles.Result {

			groupKey := p.Name + r.Group

			attrs := []*ldap.EntryAttribute{}
			attrs = append(attrs, &ldap.EntryAttribute{Name: "cn", Values: []string{r.Group}})
			attrs = append(attrs, &ldap.EntryAttribute{Name: "description", Values: []string{fmt.Sprintf("%s via Zitadel", r.Group)}})
			attrs = append(attrs, &ldap.EntryAttribute{Name: "gidNumber", Values: []string{fmt.Sprintf("%d", z.hash(groupKey))}})
			attrs = append(attrs, &ldap.EntryAttribute{Name: "uniqueMember", Values: grantMap[groupKey]})

			if asGroupOfUniqueNames {
				attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"groupOfUniqueNames", "top"}})
			} else {
				attrs = append(attrs, &ldap.EntryAttribute{Name: "memberUid", Values: grantMap[groupKey]})
				attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"posixGroup", "top"}})
			}

			dn := fmt.Sprintf("%s=%s,%s", z.options.Backend.GroupFormat, r.Group, z.baseDN)
			if hierarchy != "" {
				dn = fmt.Sprintf("%s=%s,%s,%s", z.options.Backend.GroupFormat, r.Group, hierarchy, z.baseDN)
			}

			entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
		}

	}

	return entries, nil

}

func (z *zitadelHandler) hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func (z *zitadelHandler) addUserHierarchy(user string, hierarchy string) string {
	if hierarchy != "" {
		return fmt.Sprintf("%s=%s,%s,%s", z.options.Backend.NameFormat, user, hierarchy, z.baseDN)
	}
	return fmt.Sprintf("%s=%s,ou=users,%s", z.options.Backend.NameFormat, user, z.baseDN)
}

func (z *zitadelHandler) memberOf(userID string) ([]string, error) {
	grants, err := z.zitadelClient.ListGrants()
	if err != nil {
		return nil, err
	}

	groups := []string{}

	for _, g := range grants.Result {
		if g.UserID == userID {
			for _, rk := range g.RoleKeys {
				key := fmt.Sprintf("%s=%s,ou=groups,%s", z.options.Backend.GroupFormat, rk, z.baseDN)
				groups = append(groups, key)
			}
		}
	}

	return groups, nil
}

func (z *zitadelHandler) getUserGroupIDs(userID string) ([]int, error) {
	groups := []int{}

	grants, err := z.zitadelClient.ListGrants()
	if err != nil {
		return nil, err
	}

	for _, g := range grants.Result {
		if g.UserID == userID {
			for _, r := range g.RoleKeys {
				groups = append(groups, z.hash(g.ProjectID+r))
			}
		}
	}

	return groups, nil
}
