import { beforeEach, describe, expect, it, vi } from 'vitest';

const app = { name: 'app' };
const authInstance = {
	authStateReady: vi.fn(async () => {}),
	currentUser: null as null | { getIdToken: (force?: boolean) => Promise<string> }
};

const initializeApp = vi.fn((_config: Record<string, string>) => app);
const getApps = vi.fn((): unknown[] => []);
const getAuth = vi.fn((_app: unknown) => authInstance);
const connectAuthEmulator = vi.fn();
const onAuthStateChanged = vi.fn((_auth: unknown, _next: unknown) => () => {});

vi.mock('firebase/app', () => ({ initializeApp, getApps }));
vi.mock('firebase/auth', () => ({
	getAuth,
	connectAuthEmulator,
	onAuthStateChanged,
	GoogleAuthProvider: class {},
	signInWithPopup: vi.fn(),
	signOut: vi.fn()
}));

const { createFirebaseAuth } = await import('../src/lib/auth/firebase');
const { normalizeEmulatorHost } = await import('../src/lib/auth/emulator-host');

const realConfig = {
	apiKey: 'real-key',
	authDomain: 'kumelen.firebaseapp.com',
	projectId: 'kumelen',
	appId: '1:1:web:1',
	messagingSenderId: '1'
};

/** Fuerza la carga del SDK y espera a que termine de inicializarse. */
async function load(options: Parameters<typeof createFirebaseAuth>[0]) {
	const provider = createFirebaseAuth(options);
	await provider.getToken();
	return provider;
}

beforeEach(() => {
	vi.clearAllMocks();
	getApps.mockReturnValue([]);
	authInstance.currentUser = null;
});

describe('createFirebaseAuth y el emulador', () => {
	it('con el host vacío no se conecta al emulador y usa la configuración tal cual', async () => {
		await load({ config: realConfig, emulatorHost: '' });

		expect(connectAuthEmulator).not.toHaveBeenCalled();
		expect(initializeApp).toHaveBeenCalledWith(realConfig);
	});

	it('con el host definido se conecta a esa dirección, sin avisos de consola', async () => {
		await load({ config: realConfig, emulatorHost: 'localhost:9099' });

		expect(connectAuthEmulator).toHaveBeenCalledOnce();
		expect(connectAuthEmulator).toHaveBeenCalledWith(authInstance, 'http://localhost:9099', {
			disableWarnings: true
		});
	});

	it('se conecta antes de que se use Auth (antes de escuchar la sesión)', async () => {
		const order: string[] = [];
		connectAuthEmulator.mockImplementation(() => order.push('connect'));
		onAuthStateChanged.mockImplementation(() => {
			order.push('listen');
			return () => {};
		});

		const provider = createFirebaseAuth({ config: realConfig, emulatorHost: 'localhost:9099' });
		provider.subscribe(() => {});
		await provider.getToken();

		expect(order).toEqual(['connect', 'listen']);
	});

	it('con el emulador completa la clave y el proyecto si faltan, y respeta los que vienen', async () => {
		await load({ config: { ...realConfig, apiKey: '', projectId: '' }, emulatorHost: 'localhost:9099' });
		expect(initializeApp).toHaveBeenLastCalledWith(
			expect.objectContaining({ apiKey: 'demo-api-key', projectId: 'demo-kumelen' })
		);

		initializeApp.mockClear();
		await load({ config: { ...realConfig, apiKey: 'mi-clave', projectId: 'demo-otro' }, emulatorHost: 'localhost:9099' });
		expect(initializeApp).toHaveBeenLastCalledWith(
			expect.objectContaining({ apiKey: 'mi-clave', projectId: 'demo-otro' })
		);
	});

	it('sin emulador no inventa valores para una configuración vacía', async () => {
		const empty = { ...realConfig, apiKey: '', projectId: '' };

		await load({ config: empty, emulatorHost: '' });

		expect(initializeApp).toHaveBeenCalledWith(empty);
	});

	it('entrega un token del usuario actual', async () => {
		const getIdToken = vi.fn(async () => 'id-token');
		authInstance.currentUser = { getIdToken };
		const provider = createFirebaseAuth({ config: realConfig, emulatorHost: 'localhost:9099' });

		expect(await provider.getToken(true)).toBe('id-token');
		expect(getIdToken).toHaveBeenCalledWith(true);
	});
});

describe('normalizeEmulatorHost', () => {
	it.each([
		['', ''],
		['   ', ''],
		['localhost:9099', 'localhost:9099'],
		[' localhost:9099 ', 'localhost:9099'],
		['http://localhost:9099', 'localhost:9099'],
		['http://localhost:9099/', 'localhost:9099'],
		['HTTPS://emulador.local:9099//', 'emulador.local:9099']
	])('"%s" → "%s"', (input, expected) => {
		expect(normalizeEmulatorHost(input)).toBe(expected);
	});
});
