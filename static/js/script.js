// ============================================
// GROUPIE TRACKER - JAVASCRIPT
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initThemeToggle();
    initSearchSuggestions();
});

// ============================================
// TOGGLE MODE SOMBRE
// ============================================

function initThemeToggle() {
    const themeToggle = document.getElementById('theme-toggle');
    const themeIcon = document.getElementById('theme-icon');
    
    if (!themeToggle || !themeIcon) return;

    // Vérifier la préférence sauvegardée ou utiliser le mode système
    const savedTheme = localStorage.getItem('theme');
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const currentTheme = savedTheme || (systemPrefersDark ? 'dark' : 'light');

    // Appliquer le thème
    document.documentElement.setAttribute('data-theme', currentTheme);
    updateThemeIcon(themeIcon, currentTheme);

    // Toggle au clic
    themeToggle.addEventListener('click', () => {
        const current = document.documentElement.getAttribute('data-theme');
        const newTheme = current === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
        updateThemeIcon(themeIcon, newTheme);
    });

    // Écouter les changements de préférence système
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (!localStorage.getItem('theme')) {
            const newTheme = e.matches ? 'dark' : 'light';
            document.documentElement.setAttribute('data-theme', newTheme);
            updateThemeIcon(themeIcon, newTheme);
        }
    });
}

function updateThemeIcon(icon, theme) {
    icon.textContent = theme === 'dark' ? '☀️' : '🌙';
}

// ============================================
// SUGGESTIONS DE RECHERCHE
// ============================================

function initSearchSuggestions() {
    const searchInput = document.getElementById('search-input');
    const suggestionsBox = document.getElementById('suggestions-box');

    if (!searchInput || !suggestionsBox) return;

    let debounceTimer;
    let selectedIndex = -1;

    searchInput.addEventListener('input', async function() {
        const query = this.value.trim();

        clearTimeout(debounceTimer);
        
        if (query.length < 2) {
            hideSuggestions();
            return;
        }

        debounceTimer = setTimeout(async () => {
            await fetchSuggestions(query);
        }, 300);
    });

    searchInput.addEventListener('keydown', (e) => {
        const items = suggestionsBox.querySelectorAll('.suggestion-item');
        
        if (items.length === 0) return;

        switch(e.key) {
            case 'ArrowDown':
                e.preventDefault();
                selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
                updateSelection(items);
                break;
            case 'ArrowUp':
                e.preventDefault();
                selectedIndex = Math.max(selectedIndex - 1, -1);
                updateSelection(items);
                break;
            case 'Enter':
                e.preventDefault();
                if (selectedIndex >= 0 && items[selectedIndex]) {
                    items[selectedIndex].click();
                }
                break;
            case 'Escape':
                hideSuggestions();
                break;
        }
    });

    document.addEventListener('click', (e) => {
        if (!searchInput.contains(e.target) && !suggestionsBox.contains(e.target)) {
            hideSuggestions();
        }
    });

    async function fetchSuggestions(query) {
        try {
            suggestionsBox.innerHTML = '<div class="suggestion-item" style="text-align: center;"><div class="loading" style="margin: 0 auto;"></div></div>';
            suggestionsBox.style.display = 'block';

            const response = await fetch(`/suggestions?q=${encodeURIComponent(query)}`);
            const data = await response.json();

            suggestionsBox.innerHTML = '';
            selectedIndex = -1;

            if (data && data.length > 0) {
                suggestionsBox.style.display = 'block';
                
                data.forEach((item, index) => {
                    const div = document.createElement('div');
                    div.classList.add('suggestion-item');
                    div.setAttribute('data-index', index);
                    
                    const text = item.name || item.Name || item.text;
                    const id = item.id || item.ID;

                    div.textContent = text;
                    
                    div.addEventListener('click', () => {
                        searchInput.value = text;
                        hideSuggestions();
                        
                        if (id) {
                            window.location.href = `/artist/${id}`;
                        } else {
                            searchInput.closest('form').submit();
                        }
                    });

                    div.style.opacity = '0';
                    div.style.transform = 'translateX(-10px)';
                    setTimeout(() => {
                        div.style.transition = 'all 0.3s ease-out';
                        div.style.opacity = '1';
                        div.style.transform = 'translateX(0)';
                    }, index * 50);

                    suggestionsBox.appendChild(div);
                });
            } else {
                suggestionsBox.innerHTML = '<div class="suggestion-item" style="text-align: center;">Aucun résultat trouvé</div>';
                suggestionsBox.style.display = 'block';
            }
        } catch (error) {
            console.error('Erreur suggestions:', error);
            hideSuggestions();
        }
    }

    function updateSelection(items) {
        items.forEach((item, index) => {
            if (index === selectedIndex) {
                item.style.background = 'var(--bg-tertiary)';
                item.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
            } else {
                item.style.background = '';
            }
        });
    }

    function hideSuggestions() {
        suggestionsBox.style.display = 'none';
        selectedIndex = -1;
    }
}
