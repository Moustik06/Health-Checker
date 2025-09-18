# Étude de Cas : Conception et Optimisation d'un Système Distribué Haute Performance en Go

Ce projet est une étude de cas pratique sur la conception, la construction et l'optimisation itérative d'un système de "health checking" distribué. L'objectif était de partir d'une architecture de base et de la faire évoluer pour répondre aux défis du monde réel en matière de concurrence, de robustesse, de scalabilité et de performance.

Le parcours s'est concentré sur une approche SRE (Site Reliability Engineering) : **mesurer, identifier les goulots d'étranglement, et améliorer l'architecture par itérations successives basées sur des données concrètes.**

## Sommaire
1.  [Architecture Finale](#architecture-finale)
2.  [Fonctionnalités Clés](#fonctionnalités-clés)
3.  [Le Parcours Itératif : De 65 RPS à 970 RPS](#le-parcours-itératif--de-65-rps-à-970-rps)
    *   [V1 : La Fondation Robuste (~65 RPS)](#v1--la-fondation-robuste---65-rps)
    *   [V2 : L'Optimisation par le Cache (~970 RPS)](#v2--loptimisation-par-le-cache---970-rps)
4.  [Démonstration de Compétences](#démonstration-de-compétences)
5.  [Comment Lancer le Projet](#comment-lancer-le-projet)
6.  [Stack Technique](#stack-technique)

---

## Architecture Finale

Le système est une architecture microservices entièrement conteneurisée, composée des services suivants :

*   **API Gateway (`api-gateway`)**: Point d'entrée REST unique. Valide et transmet les requêtes.
*   **Worker (`health-worker`)**: Service gRPC stateless responsable de l'exécution des vérifications HTTP.
*   **Cache Distribué (`redis`)**: Couche de cache partagée pour dédupliquer les requêtes et améliorer radicalement la performance.
*   **Stack d'Observabilité**:
    *   **Prometheus**: Collecte les métriques système et applicatives.
    *   **Grafana**: Visualise les métriques via des dashboards.
    *   **cAdvisor**: Expose les métriques des conteneurs.

---

## Fonctionnalités Clés

*   **Communication gRPC :** Contrat d'API strict et communication haute performance entre les services.
*   **Concurrence Efficace :** Utilisation d'un Worker Pool en Go pour paralléliser les tâches I/O.
*   **Robustesse par Conception :**
    *   **Graceful Shutdown** pour un arrêt sans perte de données.
    *   **Recovery Middleware** pour prévenir les crashs de la gateway.
    *   **Timeouts** stricts sur tous les appels réseau externes.
    *   **Propagation du `context.Context`** de bout en bout.
*   **Observabilité Complète :** Export de métriques Prometheus personnalisées (latence, taux d'erreur, cache hit/miss) pour une visibilité totale sur la santé et la performance du système.

---

## Le Parcours Itératif : De 65 RPS à 970 RPS

Le cœur de ce projet a été une démarche d'optimisation basée sur des tests de charge rigoureux menés avec **k6**.

### V1 : La Fondation Robuste (~65 RPS)

La première version du système a été construite en se concentrant sur la robustesse et la propreté de l'architecture.

*   **Défi :** Comment gérer efficacement des centaines de requêtes I/O simultanées sans bloquer le service ?
*   **Solution :** Implémentation d'un Worker Pool. Chaque requête gRPC déclenche un batch de jobs traités en parallèle par un nombre configurable de goroutines.
*   **Résultat des Tests de Charge :** Le système s'est avéré extrêmement robuste, avec **0% d'erreur** sous charge. Cependant, les tests ont révélé une capacité maximale d'environ **65 requêtes par seconde**.
*   **Analyse :** L'analyse des métriques a montré que le système était **"I/O Bound"**. La performance était directement limitée par la latence des réponses des serveurs externes.

### V2 : L'Optimisation par le Cache (~970 RPS)

L'analyse de la V1 a clairement montré que le goulot d'étranglement était le travail redondant. La solution était de ne jamais refaire un travail dont la réponse était déjà connue.

*   **Défi :** Comment réduire drastiquement la latence et augmenter le débit pour des charges de travail répétitives ?
*   **Solution :** Intégration de **Redis** comme couche de cache distribué. Le worker a été rendu "cache-aware" :
    1.  Vérifier d'abord Redis pour chaque URL.
    2.  Si la donnée est présente (**Cache Hit**), la servir instantanément.
    3.  Sinon (**Cache Miss**), exécuter la vérification via le Worker Pool et stocker le résultat dans Redis avec un TTL.
*   **Résultat des Tests de Charge :** La performance a été transformée. Le système a pu soutenir une charge de près de **971 requêtes par seconde**.
*   **Analyse :** La latence médiane a chuté de plusieurs secondes à **2.69 millisecondes**, et le taux de "cache hit" a dépassé **99%**. Cela a prouvé que le découplage de la latence externe via une stratégie de cache était l'optimisation la plus impactante.

---

## Démonstration de Compétences

*   **Go (Golang) :** Maîtrise de la concurrence (goroutines, canaux), des contextes, et de l'écosystème (gRPC, Prometheus client).
*   **Architecture Microservices :** Conception et implémentation d'un système distribué, découplé et scalable.
*   **Ingénierie de la Performance et de la Fiabilité (SRE) :**
    *   Mise en place d'une stack d'observabilité complète.
    *   Conduite de tests de charge méthodiques pour identifier les goulots d'étranglement.
    *   Prise de décisions architecturales basées sur des données mesurées pour améliorer la performance d'un facteur x15.
*   **DevOps & Conteneurisation :** Utilisation de Docker et docker-compose pour créer un environnement de développement et de test reproductible.

---

## Comment Lancer le Projet

1.  **Prérequis :**
    *   Docker & Docker-Compose
    *   Go (pour la génération du code Protobuf)
    *   `protoc` (le compilateur Protobuf)

2.  **Installation :**
    ```bash
    # Clonez le dépôt
    git clone https://github.com/votre-nom/healthchecker-solo.git
    cd healthchecker-solo
    
    # Générez le code gRPC (à faire une seule fois)
    protoc --proto_path=proto --go_out=. --go-grpc_out=. proto/health.proto
    ```

3.  **Lancement de l'environnement :**
    ```bash
    # Lancez tous les services en arrière-plan
    docker-compose up --build -d
    ```

4.  **Points d'accès :**
    *   **API Gateway :** `http://localhost:8080`
    *   **Prometheus :** `http://localhost:9093`
    *   **Grafana :** `http://localhost:3000` (login: admin/admin)

5.  **Exemple de Requête :**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
    -d '{"urls": ["https://google.com", "https://github.com"]}' \
    http://localhost:8080/check
    ```
    Ou depuis K6 si installé :
     ```bash
    k6 run ./tests/crash_test.js
    ```

---

## Stack Technique

*   **Langage :** Go (Golang)
*   **Communication :** gRPC, Protobuf
*   **Frameworks :** `net/http`, `go-redis`
*   **Base de Données / Cache :** Redis
*   **Observabilité :** Prometheus, Grafana, cAdvisor
*   **Conteneurisation :** Docker, Docker-Compose
*   **Test de Charge :** k6

---