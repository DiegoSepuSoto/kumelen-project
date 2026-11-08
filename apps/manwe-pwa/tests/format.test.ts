import { describe, expect, it } from 'vitest';
import { firstName, resolveUrl } from '../src/lib/format';

describe('firstName', () => {
	it('toma el primer nombre', () => {
		expect(firstName('Diego Sepúlveda Soto')).toBe('Diego');
	});

	it('ignora espacios sobrantes', () => {
		expect(firstName('  Camila   Rojas ')).toBe('Camila');
	});

	it('usa un saludo genérico si no hay nombre', () => {
		expect(firstName('')).toBe('Estudiante');
		expect(firstName(undefined)).toBe('Estudiante');
		expect(firstName(null)).toBe('Estudiante');
	});
});

describe('resolveUrl', () => {
	it('antepone la base a las rutas relativas', () => {
		expect(resolveUrl('/uploads/a.png', 'https://cms.test')).toBe('https://cms.test/uploads/a.png');
		expect(resolveUrl('uploads/a.png', 'https://cms.test')).toBe('https://cms.test/uploads/a.png');
	});

	it('no modifica las URL absolutas ni las data:', () => {
		expect(resolveUrl('https://x.test/a.png', 'https://cms.test')).toBe('https://x.test/a.png');
		expect(resolveUrl('//x.test/a.png', 'https://cms.test')).toBe('//x.test/a.png');
		expect(resolveUrl('data:image/png;base64,AAA', 'https://cms.test')).toBe('data:image/png;base64,AAA');
	});
});
