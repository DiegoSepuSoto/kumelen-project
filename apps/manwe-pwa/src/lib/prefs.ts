const ONBOARDING_KEY = 'manwe:showOnboarding';

/** Preferencias locales (equivalen a UserPreferences de manwe, sin los datos de sesión). */
export function getShowOnboarding(): boolean {
	try {
		return localStorage.getItem(ONBOARDING_KEY) !== 'false';
	} catch {
		return true;
	}
}

export function setShowOnboarding(show: boolean): void {
	try {
		localStorage.setItem(ONBOARDING_KEY, String(show));
	} catch {
		// Sin almacenamiento el onboarding volverá a mostrarse; no es grave.
	}
}
