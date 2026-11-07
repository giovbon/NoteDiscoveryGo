// Alpine component factory for a UI panel.
//
// The app still has one shared state object (state.js + the feature slices).
// Instead of duplicating that object, a panel gets a lightweight facade whose
// reads/writes delegate to the root component. That gives every panel its own
// x-data boundary (own lifecycle, and a place for local state later) while
// keeping `this` inside the existing slice methods bound to the root, exactly
// as before.
//
// The facade is built in init() because Alpine only calls init() after the
// x-data object has been wrapped in its reactive proxy and the root has been
// registered by bootstrap.init().
import { getAppRoot } from './root.js';

export function createPanelScope() {
    return {
        init() {
            const root = getAppRoot();
            if (!root) {
                throw new Error('panel scope initialized before the app root');
            }

            for (const key of Object.keys(root)) {
                if (key in this) continue; // keep Alpine magics and future local state

                if (typeof root[key] === 'function') {
                    Object.defineProperty(this, key, {
                        get: () => root[key].bind(root),
                        enumerable: true,
                        configurable: true,
                    });
                } else {
                    Object.defineProperty(this, key, {
                        get: () => root[key],
                        set: (value) => { root[key] = value; },
                        enumerable: true,
                        configurable: true,
                    });
                }
            }
        },
    };
}
