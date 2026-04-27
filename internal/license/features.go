// Package license defines feature flags used throughout Pulse.
package license

// Feature constants represent optional capabilities exposed by Pulse.
const (
	FeatureAIPatrol     = "ai_patrol"     // Background AI health monitoring
	FeatureAIAlerts     = "ai_alerts"     // AI analysis when alerts fire
	FeatureAIAutoFix    = "ai_autofix"    // Automatic remediation
	FeatureKubernetesAI = "kubernetes_ai" // AI analysis of K8s

	FeatureAgentProfiles = "agent_profiles" // Centralized agent configuration profiles

	FeatureUpdateAlerts = "update_alerts" // Alerts for pending container/package updates

	FeatureRBAC              = "rbac"               // Role-Based Access Control
	FeatureAuditLogging      = "audit_logging"      // Persistent audit logs with signing
	FeatureSSO               = "sso"                // OIDC/SSO authentication
	FeatureAdvancedSSO       = "advanced_sso"       // SAML, multi-provider, role mapping
	FeatureAdvancedReporting = "advanced_reporting" // PDF/CSV reporting engine
	FeatureLongTermMetrics   = "long_term_metrics"  // Extended historical metrics

	FeatureMultiUser   = "multi_user"
	FeatureWhiteLabel  = "white_label"
	FeatureMultiTenant = "multi_tenant"
	FeatureUnlimited   = "unlimited"
)

// Tier represents a license tier.
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierProAnnual  Tier = "pro_annual"
	TierLifetime   Tier = "lifetime"
	TierMSP        Tier = "msp"
	TierEnterprise Tier = "enterprise"
)

var allFeatures = []string{
	FeatureAIPatrol,
	FeatureAIAlerts,
	FeatureAIAutoFix,
	FeatureKubernetesAI,
	FeatureAgentProfiles,
	FeatureUpdateAlerts,
	FeatureSSO,
	FeatureAdvancedSSO,
	FeatureRBAC,
	FeatureAuditLogging,
	FeatureAdvancedReporting,
	FeatureLongTermMetrics,
	FeatureMultiUser,
	FeatureWhiteLabel,
	FeatureMultiTenant,
	FeatureUnlimited,
}

// TierFeatures maps each tier to its included features. Paid tiers are kept for
// compatibility with existing persisted licenses, but all features are available
// in the default/free tier as well.
var TierFeatures = map[Tier][]string{
	TierFree:       allFeatures,
	TierPro:        allFeatures,
	TierProAnnual:  allFeatures,
	TierLifetime:   allFeatures,
	TierMSP:        allFeatures,
	TierEnterprise: allFeatures,
}

// TierHasFeature checks if a tier includes a specific feature.
func TierHasFeature(tier Tier, feature string) bool {
	features, ok := TierFeatures[tier]
	if !ok {
		return false
	}
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetTierDisplayName returns a human-readable name for the tier.
func GetTierDisplayName(tier Tier) string {
	switch tier {
	case TierFree:
		return "Free"
	case TierPro:
		return "Pro Intelligence (Monthly)"
	case TierProAnnual:
		return "Pro Intelligence (Annual)"
	case TierLifetime:
		return "Pro Intelligence (Lifetime)"
	case TierMSP:
		return "MSP"
	case TierEnterprise:
		return "Enterprise"
	default:
		return "Unknown"
	}
}

// GetFeatureDisplayName returns a human-readable name for a feature.
func GetFeatureDisplayName(feature string) string {
	switch feature {
	case FeatureAIPatrol:
		return "Pulse Patrol (Background Health Checks)"
	case FeatureAIAlerts:
		return "Alert Analysis"
	case FeatureAIAutoFix:
		return "Pulse Patrol Auto-Fix"
	case FeatureKubernetesAI:
		return "Kubernetes Analysis"
	case FeatureUpdateAlerts:
		return "Update Alerts (Container/Package Updates)"
	case FeatureRBAC:
		return "Role-Based Access Control (RBAC)"
	case FeatureMultiUser:
		return "Multi-User Mode"
	case FeatureWhiteLabel:
		return "White-Label Branding"
	case FeatureMultiTenant:
		return "Multi-Tenant Mode"
	case FeatureUnlimited:
		return "Unlimited Instances"
	case FeatureAgentProfiles:
		return "Centralized Agent Profiles"
	case FeatureAuditLogging:
		return "Enterprise Audit Logging"
	case FeatureSSO:
		return "Basic SSO (OIDC)"
	case FeatureAdvancedSSO:
		return "Advanced SSO (SAML/Multi-Provider)"
	case FeatureAdvancedReporting:
		return "Advanced Infrastructure Reporting (PDF/CSV)"
	case FeatureLongTermMetrics:
		return "90-Day Metric History"
	default:
		return feature
	}
}
