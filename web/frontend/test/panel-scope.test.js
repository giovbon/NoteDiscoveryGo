import assert from 'node:assert/strict';
import { test, afterEach } from 'node:test';

import { createPanelScope } from '../src/app/panel-scope.js';
import { setAppRoot, getAppRoot } from '../src/app/root.js';

// ---------------------------------------------------------------------------
// Reset the app root after each test so that "init throws" works when
// no setAppRoot() has been called in the current test.
// ---------------------------------------------------------------------------
afterEach(() => {
    setAppRoot(null);
});

// ---------------------------------------------------------------------------
// Helper: creates a panel whose init() delegates to the given root.
// ---------------------------------------------------------------------------
function freshPanel(root) {
    setAppRoot(root);
    const panel = createPanelScope();
    const self = {};
    panel.init.call(self);
    return self;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
test('panel scope reads and writes delegate to the root', () => {
    const root = { count: 1, name: 'myroot' };
    const self = freshPanel(root);

    assert.equal(self.count, 1);
    assert.equal(self.name, 'myroot');

    self.count = 5;
    assert.equal(root.count, 5);
});

test('panel scope methods run with `this` bound to the root', () => {
    const root = {
        count: 1,
        increment() { this.count += 1; },
        read() { return this.count; },
    };
    const self = freshPanel(root);

    self.increment();
    assert.equal(root.count, 2);
    assert.equal(self.read(), 2);
});

test('panel scope keeps keys it already owns (Alpine magics/local state)', () => {
    const root = { value: 'root' };
    setAppRoot(root);
    const panel = createPanelScope();
    const self = { $el: 'magic', value: 'local' };
    panel.init.call(self);

    assert.equal(self.$el, 'magic');
    assert.equal(self.value, 'local');
    assert.equal(root.value, 'root');
});

test('panel scope delegates functions with `this` binding correctly', () => {
    const root = {
        counter: 10,
        add(n) { this.counter += n; },
        get() { return this.counter; },
    };
    const self = freshPanel(root);

    assert.equal(self.get(), 10);
    self.add(5);
    assert.equal(root.counter, 15);
    assert.equal(self.get(), 15);
});

test('panel scope delegates property access for nested objects', () => {
    const root = {
        settings: { theme: 'dark', fontSize: 14 },
    };
    const self = freshPanel(root);

    assert.equal(self.settings.theme, 'dark');
    assert.equal(self.settings.fontSize, 14);

    self.settings.theme = 'light';
    assert.equal(root.settings.theme, 'light');
});

test('panel scope read writes update the root when setting primitives', () => {
    const root = { flag: true, label: 'initial' };
    const self = freshPanel(root);

    assert.equal(self.flag, true);
    assert.equal(self.label, 'initial');

    self.flag = false;
    self.label = 'changed';

    assert.equal(root.flag, false);
    assert.equal(root.label, 'changed');
});

test('panel scope getAppRoot returns the stored root', () => {
    const root = { value: 42 };
    setAppRoot(root);

    assert.strictEqual(getAppRoot(), root);
});

test('panel scope multiple panels share the same root', () => {
    const root = { shared: 'value' };
    setAppRoot(root);

    const self1 = freshPanel(root);
    const self2 = freshPanel(root);

    assert.equal(self1.shared, 'value');
    assert.equal(self2.shared, 'value');

    self1.shared = 'modified';
    assert.equal(self2.shared, 'modified');
    assert.equal(root.shared, 'modified');
});

test('panel scope init throws when no root has been set', () => {
    const panel = createPanelScope();
    const self = {};

    assert.throws(() => panel.init.call(self), /panel scope initialized before the app root/);
});

test('panel scope delegates computed-like properties via get/set', () => {
    const root = {
        _count: 0,
        get count() { return this._count; },
        set count(v) { this._count = v; },
    };
    const self = freshPanel(root);

    assert.equal(self.count, 0);
    self.count = 7;
    assert.equal(root._count, 7);
    assert.equal(self.count, 7);
});

test('panel scope multiple property writes are idempotent', () => {
    const root = { a: 1, b: 2, c: 3 };
    const self = freshPanel(root);

    self.a = 10;
    self.b = 20;
    self.c = 30;

    assert.equal(root.a, 10);
    assert.equal(root.b, 20);
    assert.equal(root.c, 30);
    assert.equal(self.a, 10);
    assert.equal(self.b, 20);
    assert.equal(self.c, 30);
});