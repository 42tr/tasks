export async function apiFetch(url, options = {}) {
    const response = await fetch(url, options);

    if (response.redirected) {
        window.location.href = response.url;
        return new Promise(() => {}); // Stop execution
    }

    if (!response.ok) {
        const error = new Error(`HTTP error! status: ${response.status}`);
        error.response = response;
        try {
            error.body = await response.text();
        } catch (e) {
            // ignore
        }
        throw error;
    }

    if (response.status === 204 || response.headers.get('Content-Length') === '0') {
        // No content
        return null;
    }

    const text = await response.text();
    try {
        return JSON.parse(text);
    } catch (e) {
        return text;
    }
}
