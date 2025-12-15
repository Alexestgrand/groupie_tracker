document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const suggestionsBox = document.getElementById('suggestions-box');

    if (searchInput) {
        searchInput.addEventListener('input', async function() {
            const query = this.value;

            if (query.length < 2) { // On attend au moins 2 caractères
                suggestionsBox.style.display = 'none';
                return;
            }

            try {
                // Appel à ta route /suggestions
                const response = await fetch(`/suggestions?q=${encodeURIComponent(query)}`);
                const data = await response.json();

                suggestionsBox.innerHTML = '';

                // Vérification si on a des résultats
                // Ton API renvoie probablement soit ["Nom1", "Nom2"] soit [{"Name": "Nom1", ...}]
                if (data && data.length > 0) {
                    suggestionsBox.style.display = 'block';
                    
                    data.forEach(item => {
                        const div = document.createElement('div');
                        div.classList.add('suggestion-item');
                        
                        // Adaptation selon si ton utils.GetSuggestions renvoie des strings ou des objets
                        const text = typeof item === 'string' ? item : (item.Name || item.text);
                        const id = item.ID || item.id;

                        div.textContent = text;
                        
                        div.addEventListener('click', () => {
                            searchInput.value = text;
                            suggestionsBox.style.display = 'none';
                            // Si on a un ID, on peut rediriger directement
                            if (id) {
                                window.location.href = `/artist/${id}`;
                            } else {
                                // Sinon on lance la recherche standard
                                searchInput.closest('form').submit();
                            }
                        });
                        suggestionsBox.appendChild(div);
                    });
                } else {
                    suggestionsBox.style.display = 'none';
                }
            } catch (error) {
                console.error('Erreur suggestions:', error);
            }
        });

        // Fermer les suggestions au clic ailleurs
        document.addEventListener('click', (e) => {
            if (e.target !== searchInput && e.target !== suggestionsBox) {
                suggestionsBox.style.display = 'none';
            }
        });
    }
});
