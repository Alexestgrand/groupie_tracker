// ============================================
// GROUPIE TRACKER - JAVASCRIPT
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initThemeToggle();
    initSearchSuggestions();
    initFilters();
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

// ============================================
// GESTION DES FILTRES
// ============================================

function initFilters() {
    const filtersToggle = document.querySelector('.filters-toggle');
    const filtersContent = document.querySelector('.filters-content');
    const filtersToggleIcon = document.querySelector('.filters-toggle-icon');
    
    if (!filtersToggle || !filtersContent) return;
    
    // Vérifier s'il y a des filtres actifs
    const hasActiveFilters = document.querySelector('input[type="number"][value]:not([value=""]), input[type="text"][value]:not([value=""]), input[type="checkbox"]:checked, select option:checked');
    
    // Par défaut, les filtres sont ouverts (affichés)
    // On peut les fermer si aucun filtre n'est actif
    let isOpen = true;
    
    if (!hasActiveFilters) {
        // Si pas de filtres actifs, on peut commencer fermé
        // Mais pour l'instant, on les laisse ouverts pour faciliter l'accès
        isOpen = true;
    }
    
    // Mettre à jour l'icône selon l'état initial
    if (filtersToggleIcon) {
        filtersToggleIcon.textContent = isOpen ? '▲' : '▼';
    }
    
    if (!isOpen) {
        filtersContent.classList.add('filters-closed');
    }
    
    // Toggle au clic
    filtersToggle.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        
        const isCurrentlyOpen = !filtersContent.classList.contains('filters-closed');
        
        if (isCurrentlyOpen) {
            filtersContent.classList.add('filters-closed');
            if (filtersToggleIcon) filtersToggleIcon.textContent = '▼';
        } else {
            filtersContent.classList.remove('filters-closed');
            if (filtersToggleIcon) filtersToggleIcon.textContent = '▲';
        }
    });
    
    // Améliorer l'UX des checkboxes
    const checkboxes = document.querySelectorAll('.filter-checkbox input[type="checkbox"]');
    checkboxes.forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            // Animation visuelle
            const label = this.closest('.filter-checkbox');
            if (this.checked) {
                label.style.transform = 'scale(1.05)';
                setTimeout(() => {
                    label.style.transform = '';
                }, 200);
            }
        });
    });
}
