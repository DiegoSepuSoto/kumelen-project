import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createMockAuth } from '../src/lib/auth/mock';
import { getShowOnboarding, setShowOnboarding } from '../src/lib/prefs';

beforeEach(() => localStorage.clear());

describe('createMockAuth', () => {
	it('empieza sin sesión y sin token', async () => {
		const auth = createMockAuth();
		const onChange = vi.fn();

		auth.subscribe(onChange);

		expect(onChange).toHaveBeenCalledWith(null);
		expect(await auth.getToken()).toBeNull();
	});

	it('al ingresar notifica al usuario y entrega un token', async () => {
		const auth = createMockAuth();
		const onChange = vi.fn();
		auth.subscribe(onChange);

		const user = await auth.signIn();

		expect(onChange).toHaveBeenLastCalledWith(user);
		expect(await auth.getToken()).toBe('mock-token');
	});

	it('conserva la sesión entre cargas de la página', async () => {
		await createMockAuth().signIn();

		const onChange = vi.fn();
		createMockAuth().subscribe(onChange);

		expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ uid: 'mock-user' }));
	});

	it('al salir limpia la sesión', async () => {
		const auth = createMockAuth();
		await auth.signIn();
		const onChange = vi.fn();
		auth.subscribe(onChange);

		await auth.signOut();

		expect(onChange).toHaveBeenLastCalledWith(null);
		expect(await auth.getToken()).toBeNull();
	});

	it('deja de notificar a quien se desuscribe', async () => {
		const auth = createMockAuth();
		const onChange = vi.fn();
		const stop = auth.subscribe(onChange);
		stop();
		onChange.mockClear();

		await auth.signIn();

		expect(onChange).not.toHaveBeenCalled();
	});
});

describe('onboarding', () => {
	it('se muestra por defecto y deja de mostrarse tras completarlo', () => {
		expect(getShowOnboarding()).toBe(true);

		setShowOnboarding(false);

		expect(getShowOnboarding()).toBe(false);
	});
});
