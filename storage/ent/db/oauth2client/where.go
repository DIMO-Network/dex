// Code generated by entc, DO NOT EDIT.

package oauth2client

import (
	"entgo.io/ent/dialect/sql"
	"github.com/dexidp/dex/storage/ent/db/predicate"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// Secret applies equality check predicate on the "secret" field. It's identical to SecretEQ.
func Secret(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldSecret), v))
	})
}

// Public applies equality check predicate on the "public" field. It's identical to PublicEQ.
func Public(v bool) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldPublic), v))
	})
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldName), v))
	})
}

// LogoURL applies equality check predicate on the "logo_url" field. It's identical to LogoURLEQ.
func LogoURL(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLogoURL), v))
	})
}

// SecretEQ applies the EQ predicate on the "secret" field.
func SecretEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldSecret), v))
	})
}

// SecretNEQ applies the NEQ predicate on the "secret" field.
func SecretNEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldSecret), v))
	})
}

// SecretIn applies the In predicate on the "secret" field.
func SecretIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldSecret), v...))
	})
}

// SecretNotIn applies the NotIn predicate on the "secret" field.
func SecretNotIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldSecret), v...))
	})
}

// SecretGT applies the GT predicate on the "secret" field.
func SecretGT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldSecret), v))
	})
}

// SecretGTE applies the GTE predicate on the "secret" field.
func SecretGTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldSecret), v))
	})
}

// SecretLT applies the LT predicate on the "secret" field.
func SecretLT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldSecret), v))
	})
}

// SecretLTE applies the LTE predicate on the "secret" field.
func SecretLTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldSecret), v))
	})
}

// SecretContains applies the Contains predicate on the "secret" field.
func SecretContains(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldSecret), v))
	})
}

// SecretHasPrefix applies the HasPrefix predicate on the "secret" field.
func SecretHasPrefix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldSecret), v))
	})
}

// SecretHasSuffix applies the HasSuffix predicate on the "secret" field.
func SecretHasSuffix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldSecret), v))
	})
}

// SecretEqualFold applies the EqualFold predicate on the "secret" field.
func SecretEqualFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldSecret), v))
	})
}

// SecretContainsFold applies the ContainsFold predicate on the "secret" field.
func SecretContainsFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldSecret), v))
	})
}

// RedirectUrisIsNil applies the IsNil predicate on the "redirect_uris" field.
func RedirectUrisIsNil() predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.IsNull(s.C(FieldRedirectUris)))
	})
}

// RedirectUrisNotNil applies the NotNil predicate on the "redirect_uris" field.
func RedirectUrisNotNil() predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NotNull(s.C(FieldRedirectUris)))
	})
}

// TrustedPeersIsNil applies the IsNil predicate on the "trusted_peers" field.
func TrustedPeersIsNil() predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.IsNull(s.C(FieldTrustedPeers)))
	})
}

// TrustedPeersNotNil applies the NotNil predicate on the "trusted_peers" field.
func TrustedPeersNotNil() predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NotNull(s.C(FieldTrustedPeers)))
	})
}

// PublicEQ applies the EQ predicate on the "public" field.
func PublicEQ(v bool) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldPublic), v))
	})
}

// PublicNEQ applies the NEQ predicate on the "public" field.
func PublicNEQ(v bool) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldPublic), v))
	})
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldName), v))
	})
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldName), v))
	})
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldName), v...))
	})
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldName), v...))
	})
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldName), v))
	})
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldName), v))
	})
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldName), v))
	})
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldName), v))
	})
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldName), v))
	})
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldName), v))
	})
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldName), v))
	})
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldName), v))
	})
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldName), v))
	})
}

// LogoURLEQ applies the EQ predicate on the "logo_url" field.
func LogoURLEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLogoURL), v))
	})
}

// LogoURLNEQ applies the NEQ predicate on the "logo_url" field.
func LogoURLNEQ(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldLogoURL), v))
	})
}

// LogoURLIn applies the In predicate on the "logo_url" field.
func LogoURLIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldLogoURL), v...))
	})
}

// LogoURLNotIn applies the NotIn predicate on the "logo_url" field.
func LogoURLNotIn(vs ...string) predicate.OAuth2Client {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.OAuth2Client(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldLogoURL), v...))
	})
}

// LogoURLGT applies the GT predicate on the "logo_url" field.
func LogoURLGT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldLogoURL), v))
	})
}

// LogoURLGTE applies the GTE predicate on the "logo_url" field.
func LogoURLGTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldLogoURL), v))
	})
}

// LogoURLLT applies the LT predicate on the "logo_url" field.
func LogoURLLT(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldLogoURL), v))
	})
}

// LogoURLLTE applies the LTE predicate on the "logo_url" field.
func LogoURLLTE(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldLogoURL), v))
	})
}

// LogoURLContains applies the Contains predicate on the "logo_url" field.
func LogoURLContains(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldLogoURL), v))
	})
}

// LogoURLHasPrefix applies the HasPrefix predicate on the "logo_url" field.
func LogoURLHasPrefix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldLogoURL), v))
	})
}

// LogoURLHasSuffix applies the HasSuffix predicate on the "logo_url" field.
func LogoURLHasSuffix(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldLogoURL), v))
	})
}

// LogoURLEqualFold applies the EqualFold predicate on the "logo_url" field.
func LogoURLEqualFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldLogoURL), v))
	})
}

// LogoURLContainsFold applies the ContainsFold predicate on the "logo_url" field.
func LogoURLContainsFold(v string) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldLogoURL), v))
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OAuth2Client) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OAuth2Client) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OAuth2Client) predicate.OAuth2Client {
	return predicate.OAuth2Client(func(s *sql.Selector) {
		p(s.Not())
	})
}