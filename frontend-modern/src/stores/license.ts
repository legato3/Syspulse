import { createSignal, createMemo } from 'solid-js';
import { LicenseAPI, type LicenseStatus } from '@/api/license';
import { logger } from '@/utils/logger';

// Reactive signals for license status
const [licenseStatus, setLicenseStatus] = createSignal<LicenseStatus | null>(null);
const [loading, setLoading] = createSignal(false);
const [loaded, setLoaded] = createSignal(false);

const ALL_FEATURES = [
    'ai_patrol',
    'ai_alerts',
    'ai_autofix',
    'kubernetes_ai',
    'agent_profiles',
    'update_alerts',
    'sso',
    'advanced_sso',
    'rbac',
    'audit_logging',
    'advanced_reporting',
    'long_term_metrics',
    'multi_user',
    'white_label',
    'multi_tenant',
    'unlimited',
];

/**
 * Load the license status from the server.
 */
export async function loadLicenseStatus(force = false): Promise<void> {
    if (loaded() && !force) return;

    setLoading(true);
    try {
        const status = await LicenseAPI.getStatus();
        setLicenseStatus(status);
        setLoaded(true);
        logger.debug('[licenseStore] License status loaded', { tier: status.tier, valid: status.valid });
    } catch (err) {
        logger.error('[licenseStore] Failed to load license status', err);
        // Fallback to free tier on error to avoid breaking UI
        setLicenseStatus({
            valid: true,
            tier: 'free',
            is_lifetime: false,
            days_remaining: 0,
            features: ALL_FEATURES,
        });
        setLoaded(true);
    } finally {
        setLoading(false);
    }
}

/**
 * Helper retained for components that previously hid paid badges.
 */
export const isPro = createMemo(() => {
    const current = licenseStatus();
    return Boolean(current?.features.length);
});

/**
 * Check if a specific feature is enabled by the current license.
 * Free tier features (e.g., ai_patrol) are available even without a valid Pro license.
 */
export function hasFeature(feature: string): boolean {
    const current = licenseStatus();
    if (!current) return ALL_FEATURES.includes(feature);
    return current.features.includes(feature);
}

/**
 * Get the full license status.
 */
export { licenseStatus, loading as licenseLoading, loaded as licenseLoaded };
