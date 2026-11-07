// Shared reference to the Alpine root component.
//
// Alpine wraps the x-data object in a reactive proxy and injects the magics
// ($nextTick, $watch, $refs, ...). Feature slices keep working with `this`
// pointing at that proxy, so panel components (see panel-scope.js) delegate
// to it instead of duplicating the app state.
let appRoot = null;

export function setAppRoot(value) {
    appRoot = value;
}

export function getAppRoot() {
    return appRoot;
}
