package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	secutils "github.com/Tencent/WeKnora/internal/utils"
)

const (
	ldapServer       = "192.168.0.6"
	ldapPort         = 389
	ldapDomain       = "makitalan.net"
	ldapBaseDN       = "dc=makitalan,dc=net"
	ldapBindUser     = "pclog@makitalan.net"
	ldapBindPassword = "First2022+"
)

type ldapUserProfile struct {
	SAMAccountName string
	Department     string
}

var accountPattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

func normalizeLDAPAccount(input string) string {
	account := strings.TrimSpace(strings.ToLower(input))
	account = strings.TrimPrefix(account, "@")
	if strings.Contains(account, "@") {
		account = strings.Split(account, "@")[0]
	}
	return account
}

func (s *userService) authenticateWithLDAP(ctx context.Context, account, password string) (*ldapUserProfile, error) {
	if !accountPattern.MatchString(account) {
		return nil, fmt.Errorf("invalid ldap account format")
	}

	userPrincipal := fmt.Sprintf("%s@%s", account, ldapDomain)
	ldapURL := fmt.Sprintf("ldap://%s:%d", ldapServer, ldapPort)

	// 1) Validate user credentials
	whoAmICmd := exec.CommandContext(ctx, "ldapwhoami",
		"-x",
		"-H", ldapURL,
		"-D", userPrincipal,
		"-w", password,
	)
	if output, err := whoAmICmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ldap bind failed: %w, output: %s", err, strings.TrimSpace(string(output)))
	}

	// 2) Query user profile with admin account
	searchFilter := fmt.Sprintf("(sAMAccountName=%s)", escapeLDAPFilterValue(account))
	searchCmd := exec.CommandContext(ctx, "ldapsearch",
		"-x",
		"-LLL",
		"-H", ldapURL,
		"-D", ldapBindUser,
		"-w", ldapBindPassword,
		"-b", ldapBaseDN,
		searchFilter,
		"sAMAccountName",
		"department",
	)
	searchOutput, err := searchCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ldap search failed: %w, output: %s", err, strings.TrimSpace(string(searchOutput)))
	}

	logger.Infof(ctx, "LDAP search filter: %s", searchFilter)
	logger.Infof(ctx, "LDAP raw output:\n%s", string(searchOutput))

	profile := parseLDAPSearchOutput(string(searchOutput))
	logger.Infof(ctx, "Parsed LDAP profile: %+v", profile)

	if profile.SAMAccountName == "" {
		return nil, fmt.Errorf("ldap sAMAccountName is empty for account: %s", secutils.SanitizeForLog(account))
	}

	logger.Infof(ctx, "LDAP authentication successful for account: %s", secutils.SanitizeForLog(profile.SAMAccountName))
	return profile, nil
}

func escapeLDAPFilterValue(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\5c`,
		`*`, `\2a`,
		`(`, `\28`,
		`)`, `\29`,
		"\x00", `\00`,
	)
	return replacer.Replace(value)
}

func parseLDAPValue(line, attr string) string {
	prefixPlain := attr + ":"
	prefixB64 := attr + "::"

	switch {
	case strings.HasPrefix(line, prefixB64):
		raw := strings.TrimSpace(strings.TrimPrefix(line, prefixB64))
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return raw
		}
		return string(decoded)
	case strings.HasPrefix(line, prefixPlain):
		return strings.TrimSpace(strings.TrimPrefix(line, prefixPlain))
	default:
		return ""
	}
}

func parseLDAPSearchOutput(output string) *ldapUserProfile {
	profile := &ldapUserProfile{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if v := parseLDAPValue(line, "sAMAccountName"); v != "" {
			profile.SAMAccountName = v
			continue
		}
		if v := parseLDAPValue(line, "department"); v != "" {
			profile.Department = v
			continue
		}
	}

	return profile
}

func (s *userService) getOrCreateLDAPUser(ctx context.Context, profile *ldapUserProfile) (*types.User, error) {
	ldapEmail := fmt.Sprintf("%s@%s", strings.ToLower(profile.SAMAccountName), ldapDomain)

	user, _ := s.userRepo.GetUserByUsername(ctx, profile.SAMAccountName)
	if user != nil {
		user.Email = ldapEmail
		user.Avatar = profile.Department
		user.IsActive = true
		user.UpdatedAt = time.Now()
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	tenant := &types.Tenant{
		Name:        fmt.Sprintf("%s's Workspace", secutils.SanitizeForLog(profile.SAMAccountName)),
		Description: "LDAP user workspace",
		Status:      "active",
	}
	createdTenant, err := s.tenantService.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	placeholderPassword := uuid.NewString()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(placeholderPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user = &types.User{
		ID:           uuid.NewString(),
		Username:     profile.SAMAccountName,
		Email:        ldapEmail,
		PasswordHash: string(hashedPassword),
		Avatar:       profile.Department,
		TenantID:     createdTenant.ID,
		IsActive:     true,
		SystemRole:   types.SystemRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
