import matplotlib.pyplot as plt
import numpy as np

# Utilisateurs pour lesquels nous avons mesuré les temps
users = [44, 318, 387, 414, 448, 599, 600, 610]

# Temps d'exécution (en ms) mesurés pour le pipeline avec 1 module et 2 modules de calcul du score.
# Remplacez ces valeurs par vos mesures réelles.
times_1module = [2373, 3747, 3949, 2149, 5628, 3840, 3441, 4250]
times_2modules = [3493, 4123, 3798, 1530, 4778, 2905, 3515, 3884]

x = np.arange(len(users))  # emplacements sur l'axe des x pour chaque utilisateur
width = 0.35  # largeur des barres

fig, ax = plt.subplots()
rects1 = ax.bar(x - width/2, times_1module, width, label='1 module')
rects2 = ax.bar(x + width/2, times_2modules, width, label='2 modules')

# Ajout des labels, titre et légende
ax.set_ylabel('Temps d\'exécution (ms)')
ax.set_xlabel('ID Utilisateur')
ax.set_title('Comparaison des temps d\'exécution: 1 module vs 2 modules')
ax.set_xticks(x)
ax.set_xticklabels([str(u) for u in users])
ax.legend()

# Fonction pour ajouter les labels de valeur sur les barres
def autolabel(rects):
    for rect in rects:
        height = rect.get_height()
        ax.annotate(f'{height}',
                    xy=(rect.get_x() + rect.get_width()/2, height),
                    xytext=(0, 3),  # décalage en points
                    textcoords="offset points",
                    ha='center', va='bottom')

autolabel(rects1)
autolabel(rects2)

fig.tight_layout()
plt.show()
