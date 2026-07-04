#import "@preview/mmdr:0.2.2": mermaid
#import "template.typ": project

#show table.cell: set par(justify: false)

#show: doc => project(
  title: [
    INFORME DE AVANCE DE PROYECTO: MICROPHOTO
  ],
  authors: (
    "BEDREGAL PEREZ, DANIEL",
    "JARA MAMANI, MARIEL ALISSON",
    "MESTAS ZEGARRA, CHRISTIAN RAUL",
    "NOA CAMINO, YENARO JOEL",
    "SEQUEIROS CONDORI, LUIS GUSTAVO",
  ),
  course: "SISTEMAS DISTRIBUIDOS",
  group: "GRUPO 2",
  teacher: "MOLINA BARRIGA, MARIBEL",
  doc,
)

= Enlace a GitHub
https://github.com/christianmz565/microphoto

= Introducción
Este informe de avance del proyecto Microphoto registra el estado actual del desarrollo, detalla los componentes implementados y señala las tareas pendientes.

= Propósito
Este informe documenta el estado de avance técnico del proyecto al corte del 24 de junio de 2026, con el objetivo de:

- Comunicar el porcentaje de avance frente a lo planificado.
- Evidenciar los módulos entregables ya funcionales: backend distribuido (coordinador, worker, reaper), frontend web y configuración de infraestructura.
- Identificar las funcionalidades aún en desarrollo y proponer las acciones pendientes para el cierre del proyecto.

= Alcance del informe
El informe cubre el periodo comprendido entre el inicio del proyecto y el 24 de junio de 2026. Se evalúan los siguientes módulos o áreas del sistema:

- Backend en Go @go-language: coordinador HTTP, workers de procesamiento y reaper de tareas colgadas.
- Frontend web: aplicación Astro @astro-docs con componentes React e integración via SSE.
- Infraestructura: configuración Docker Compose @docker-compose y manifiestos Kubernetes/Helm.
- Definición del esquema de mensajería con Protocol Buffers @protobuf gestionado con la herramienta buf @buf-build.

El avance se calcula a partir de los módulos funcionales del repositorio.

= Descripción general

== Contexto del proyecto
Microphoto es una plataforma web de procesamiento de imágenes que emplea una arquitectura de sistemas distribuidos para dividir, procesar en paralelo y reconstruir imágenes aplicando filtros digitales. El sistema resuelve el problema de los tiempos de procesamiento secuencial elevados en imágenes de gran tamaño, distribuyendo el trabajo entre múltiples nodos worker.

El proyecto persigue los siguientes objetivos técnicos:

- Implementar un pipeline de procesamiento distribuido que fragmente automáticamente las imágenes en subtareas, las distribuya entre workers disponibles y reconstruya la imagen final.
- Ofrecer retroalimentación en tiempo real al usuario sobre el estado del procesamiento mediante Server-Sent Events (SSE).
- Escalar horizontalmente el número de workers mediante réplicas en contenedores Docker o Kubernetes.

Los filtros disponibles en la versión actual son: escala de grises (GRAYSCALE), desenfoque gaussiano (BLUR), ajuste de brillo (BRIGHTNESS) y redimensionamiento (RESIZE).

== Estado general del proyecto
El proyecto se encuentra en un avance aproximado del 50 %, con la arquitectura central completamente diseñada e implementada. Los tres servicios del backend están operativos y se comunican correctamente a través de Redis @redis-docs y MinIO @minio-go. El frontend cuenta con todas las vistas e integración SSE funcionales. La configuración de Docker Compose permite levantar el sistema completo de forma local. El despliegue en Kubernetes está parcialmente configurado, a falta de validación en un entorno real. Las pruebas automatizadas, la autenticación de usuarios y el monitoreo con Prometheus/Grafana completo son funcionalidades pendientes de implementación.

= Equipo y roles

== Roles del equipo

#table(
  columns: (1fr, 2fr, 3fr),
  align: horizon,
  table.header([Rol], [Integrantes], [Responsabilidades]),
  [Backend],
  [Bedregal Perez, Daniel \ Mestas Zegarra, Christian Raul \ Noa Camino, Yenaro Joel],
  [Diseño e implementación del servicio coordinador HTTP, orquestación de tareas, procesamiento de imágenes con bimg y libvips, lógica de slicing y reconstitución, y reaper de control de timeouts],

  [Frontend],
  [Jara Mamani, Mariel Alisson],
  [Desarrollo de la aplicación con Astro y React, integración SSE y componentes UI con shadcn/ui],

  [Infrastructure],
  [Sequeiros Condori, Luis Gustavo],
  [Configuración Docker Compose, manifiestos Kubernetes, Helm charts, Helmfile y gestión de secretos con SOPS],
)

#pagebreak()
= Backlog

== Resumen de épicas

#table(
  columns: (1fr, 3fr, 1.2fr, 1.5fr, 1fr),
  align: horizon,
  table.header([Código], [Épica], [Total HU], [HU completadas], [Avance]),
  [E-001], [Pipeline de procesamiento distribuido de imágenes], [8], [8], [100 %],
  [E-002], [Interfaz de usuario y experiencia de cliente], [5], [5], [100 %],
  [E-003], [Observabilidad y monitoreo], [3], [1], [33 %],
  [E-004], [Infraestructura y despliegue], [4], [2], [50 %],
  [E-005], [Gestión de usuarios y seguridad], [3], [0], [0 %],
)

== Historias de usuario principales
Las historias de usuario definen las necesidades y funcionalidades desde la perspectiva del usuario final:

#table(
  columns: (1fr, 2.2fr, 4.5fr, 1fr),
  align: horizon,
  table.header([Código], [Historia de usuario], [Descripción], [Estado]),
  [HU-001],
  [Subida y procesamiento de imágenes],
  [Como usuario, deseo subir una imagen y configurar filtros para obtener una imagen editada según mis necesidades.],
  [Completo],

  [HU-002],
  [Visualización de progreso en tiempo real],
  [Como usuario, deseo ver el progreso del procesamiento en tiempo real para conocer cuánto falta para que termine la edición.],
  [Completo],

  [HU-003],
  [Descarga de la imagen final],
  [Como usuario, deseo descargar la imagen procesada en mi dispositivo para poder guardarla o utilizarla fuera de la plataforma.],
  [Completo],

  [HU-004],
  [Resiliencia ante fallos del sistema],
  [Como usuario, deseo que el procesamiento de mi imagen continúe de manera automática si ocurre una falla técnica, para evitar subir el archivo nuevamente.],
  [Completo],

  [HU-005],
  [Historial local de tareas],
  [Como usuario, deseo acceder a un historial de mis procesamientos recientes en la sesión actual para recuperar y descargar resultados anteriores.],
  [Completo],

  [HU-006],
  [Página informativa de inicio],
  [Como usuario, deseo ver una pantalla de inicio explicativa para comprender rápidamente el funcionamiento y las características de la plataforma.],
  [Completo],

  [HU-007],
  [Registro e inicio de sesión],
  [Como usuario, deseo crear una cuenta e iniciar sesión de forma segura para proteger mis datos y acceder a funciones personalizadas.],
  [Pendiente],

  [HU-008],
  [Historial persistente en la cuenta],
  [Como usuario, deseo que mi historial de imágenes se guarde de manera permanente en mi cuenta para consultarlo desde cualquier dispositivo.],
  [Pendiente],
)

#pagebreak()
= Requerimientos
Los requerimientos del sistema han sido derivados a partir de las necesidades expresadas en las historias de usuario anteriores.

== Requerimientos funcionales

#table(
  columns: (1fr, 3fr, 1.2fr, 1fr),
  align: horizon,
  table.header([Código], [Requerimiento], [Estado], [Avance]),
  [RF-001], [Subida de imágenes desde el navegador vía multipart/form-data], [Completo], [100 %],
  [RF-002],
  [Fragmentación automática de la imagen en subtareas proporcionales al número de píxeles],
  [Completo],
  [100 %],

  [RF-003], [Procesamiento paralelo de fragmentos por múltiples workers], [Completo], [100 %],
  [RF-004], [Filtro escala de grises (GRAYSCALE)], [Completo], [100 %],
  [RF-005], [Filtro desenfoque gaussiano (BLUR) con radio configurable], [Completo], [100 %],
  [RF-006], [Filtro de brillo (BRIGHTNESS) con factor configurable], [Completo], [100 %],
  [RF-007], [Filtro de redimensionamiento (RESIZE) con ancho y alto configurables], [Completo], [100 %],
  [RF-008], [Reconstrucción automática de fragmentos procesados en imagen final], [Completo], [100 %],
  [RF-009], [Notificación de progreso en tiempo real al cliente via SSE], [Completo], [100 %],
  [RF-010], [Descarga del resultado procesado desde el frontend], [Completo], [100 %],
  [RF-011], [Reintento automático de tareas que exceden el tiempo máximo (reaper)], [Completo], [100 %],
  [RF-012], [Historial de tareas procesadas en el navegador], [Completo], [100 %],
  [RF-013], [Página de landing informativa con descripción del producto], [Completo], [100 %],
  [RF-014], [Autenticación y gestión de usuarios], [Pendiente], [0 %],
  [RF-015], [Almacenamiento persistente del historial de tareas en base de datos], [Pendiente], [0 %],
)

== Requerimientos no funcionales

#table(
  columns: (1fr, 3fr, 1.2fr, 1fr),
  align: horizon,
  table.header([Código], [Atributo], [Estado], [Avance]),
  [RNF-001], [Escalabilidad horizontal: los workers se pueden replicar sin modificar el código], [Completo], [100 %],
  [RNF-002], [Tolerancia a fallos: el reaper detecta workers caídos y reagenda las tareas], [Completo], [100 %],
  [RNF-003], [Contenerización: Dockerfile multi-stage para los tres servicios], [Completo], [100 %],
  [RNF-004], [Observabilidad: métricas OpenTelemetry @opentelemetry-go exportadas a Prometheus], [En curso], [60 %],
  [RNF-005], [Gestión de objetos: imágenes almacenadas en MinIO con política de ciclo de vida], [Completo], [100 %],
  [RNF-006], [Serialización eficiente: mensajería de tareas via Protocol Buffers sobre Redis], [Completo], [100 %],
  [RNF-007], [Despliegue en Kubernetes: manifiestos y Helm charts definidos], [En curso], [50 %],
  [RNF-008], [Seguridad de secretos: gestión con SOPS en el directorio kube/], [En curso], [40 %],
  [RNF-009], [Pruebas automatizadas: cobertura de pruebas unitarias e integración], [Pendiente], [0 %],
  [RNF-010], [Tiempo de respuesta: retroalimentación al usuario en menos de 2 s tras el inicio], [Completo], [100 %],
)

#pagebreak()
= Plan de implementación

== Fases

#table(
  columns: (1fr, 3fr, 1.5fr, 1.2fr, 3fr),
  align: horizon,
  table.header([Fase], [Nombre], [Duración estimada], [Estado], [Entregable visible]),
  [1],
  [Diseño de arquitectura y esquema de mensajería],
  [1 semana],
  [Completa],
  [Diagrama de arquitectura, definición Proto (jobs.proto)],

  [2],
  [Backend: Pipeline central (coordinador + worker + reaper)],
  [3 semanas],
  [Completa],
  [Servicios ejecutables, Docker Compose funcional],

  [3], [Frontend: SPA e integración SSE], [2 semanas], [Completa], [Aplicación web accesible en navegador],
  [4],
  [Infraestructura Kubernetes y observabilidad],
  [2 semanas],
  [En curso],
  [Helm charts parcialmente configurados, métricas Prometheus],

  [5],
  [Pruebas, seguridad y despliegue],
  [2 semanas],
  [Pendiente],
  [Suite de pruebas, autenticación, despliegue en clúster real],
)

== Cronograma

#table(
  columns: (1fr, 1.5fr, 1.5fr, 1.5fr, 1.5fr, 1fr),
  align: horizon,
  table.header([Fase], [Inicio previsto], [Fin previsto], [Inicio real], [Fin real], [Avance]),
  [1], [01/05/2026], [07/05/2026], [01/05/2026], [07/05/2026], [100 %],
  [2], [08/05/2026], [28/05/2026], [08/05/2026], [28/05/2026], [100 %],
  [3], [29/05/2026], [11/06/2026], [29/05/2026], [12/06/2026], [100 %],
  [4], [12/06/2026], [25/06/2026], [12/06/2026], [(en curso)], [50 %],
  [5], [26/06/2026], [10/07/2026], [—], [—], [0 %],
)

#pagebreak()
= Arquitectura propuesta
El sistema sigue un patrón productor-cola-consumidor con almacenamiento de objetos compartido. A continuación, se detalla la arquitectura global de la solución y, posteriormente, la arquitectura individual de sus componentes e infraestructura de despliegue.

== Arquitectura Global del Sistema
La arquitectura global de Microphoto orquesta la interacción entre el navegador web, el proxy de red Traefik, el coordinador, la cola de Redis, los workers de procesamiento y el almacenamiento en MinIO.

=== Estructura y Componentes Centrales
A continuación, se detalla la topología de red y los flujos de datos principales entre los diferentes componentes del sistema:

#figure(
  mermaid(
    "flowchart LR

    subgraph Frontend[\"Frontend\"]
        BROWSER[\"Navegador<br/>Astro + React\"]
    end

    subgraph Edge[\"Borde\"]
        CF[\"Cloudflare<br/>DNS y TLS\"]
        TR[\"Traefik<br/>Balanceador e Ingress\"]
    end

    subgraph Backend[\"Servicios Go\"]
        COORD[\"Coordinator<br/>Recibe uploads y emite eventos\"]
        W1[\"Worker<br/>Procesa fragmentos de imagen\"]
        W2[\"Worker<br/>Procesa fragmentos de imagen\"]
        WN[\"Worker<br/>Procesa fragmentos de imagen\"]
        REAP[\"Reaper<br/>Reintenta trabajos colgados\"]
    end

    subgraph State[\"Estado\"]
        REDIS[(\"Redis<br/>Cola y pubsub de progreso\")]
        MINIO[(\"MinIO<br/>Almacenamiento de imagenes\")]
    end

    subgraph Observabilidad[\"Observabilidad\"]
        PROM[\"Prometheus<br/>Recolecta metricas\"]
    end

    BROWSER -->|HTTPS y SSE| CF
    CF --> TR
    TR -->|/api| COORD

    COORD -->|Encolar trabajo| REDIS
    W1 -->|Tomar trabajo y reportar| REDIS
    W2 -->|Tomar trabajo y reportar| REDIS
    WN -->|Tomar trabajo y reportar| REDIS
    REAP -->|Escanear y reintentar| REDIS

    COORD -->|Guardar original| MINIO
    W1 -->|Leer y escribir fragmentos| MINIO
    W2 -->|Leer y escribir fragmentos| MINIO
    WN -->|Leer y escribir fragmentos| MINIO

    W1 -.Comparten cola.- W2
    W1 -.Comparten cola.- WN
    W2 -.Comparten cola.- WN

    REDIS -.Eventos de progreso.- COORD
    COORD -->|Stream SSE| TR
    TR --> CF
    CF --> BROWSER

    PROM -->|Scrape| COORD
    PROM -->|Scrape| W1
    PROM -->|Scrape| W2
    PROM -->|Scrape| WN
    PROM -->|Scrape| REAP

    classDef fe fill:#581c87,stroke:#a855f7,color:#f3e8ff
    classDef edge fill:#0c4a6e,stroke:#0ea5e9,color:#e0f2fe
    classDef app fill:#1e3a8a,stroke:#3b82f6,color:#dbeafe
    classDef work fill:#14532d,stroke:#22c55e,color:#dcfce7
    classDef state fill:#7c2d12,stroke:#fbbf24,color:#fef3c7
    classDef obs fill:#0f766e,stroke:#14b8a6,color:#ccfbf1

    class BROWSER fe
    class CF,TR edge
    class COORD,REAP app
    class W1,W2,WN work
    class REDIS,MINIO state
    class PROM obs",
  ),
  caption: [Diagrama de Arquitectura Global de Microphoto],
) <fig-architecture>

=== Ciclo de Vida y Secuencia
El ciclo de vida de una tarea de procesamiento, desde la subida de la imagen y creación del trabajo de segmentación hasta la reconstrucción y entrega final, se describe cronológicamente en el diagrama de secuencia de la @fig-sequence.

#figure(
  box(
    width: 65%,
    mermaid(
      "%%{init: {\"theme\": \"default\", \"themeVariables\": {\"fontSize\": \"20px\", \"fontFamily\": \"Arial\"}, \"sequence\": {\"actorMargin\": 80, \"messageMargin\": 50, \"boxMargin\": 20, \"noteMargin\": 20, \"actorFontSize\": \"20px\", \"messageFontSize\": \"18px\", \"noteFontSize\": \"18px\"}}}%%
sequenceDiagram
    autonumber
    actor U as Usuario
    participant C as Coordinator
    participant R as Redis
    participant S as MinIO
    participant W as Worker

    note over U,C: 1 · Subida y creacion de tarea
    U->>C: POST /api/process con imagen
    C->>S: guarda original
    C->>R: encola trabajo de corte
    C-->>U: responde con id de tarea

    note over C,R: 2 · Suscripcion a progreso
    C->>R: suscribe a canal de progreso

    note over R,W: 3 · Corte de imagen
    R-->>W: asigna trabajo de corte
    W->>S: guarda fragmentos
    W->>R: inicializa contadores
    W->>R: encola trabajos de proceso

    note over R,W: 4 · Procesamiento
    R-->>W: asigna fragmento
    W->>S: lee fragmento
    W->>W: aplica filtro
    W->>S: guarda resultado
    W->>R: reporta progreso
    R-->>C: emite evento de progreso

    note over R,W: 5 · Reconstruccion
    W->>R: toma cerrojo de reconstruccion
    W->>R: encola reconstruccion
    R-->>W: asigna reconstruccion
    W->>S: lee fragmentos finales
    W->>W: compone imagen final
    W->>S: guarda resultado final
    W->>R: emite evento completado
    R-->>C: notifica finalizacion

    note over U,C: 6 · Descarga y visualizacion
    U->>C: GET /result
    C->>S: obtiene imagen final
    C-->>U: entrega resultado",
    ),
  ),
  caption: [Diagrama de Secuencia Global del Ciclo de Vida de Procesamiento],
) <fig-sequence>

=== Decisiones arquitectónicas clave
- Cola fiable con `BLMOVE`: el worker mueve atómicamente un job de `{"global"}:queue` a `{"global"}:in_progress:{workerID}`, de modo que si el worker falla, el reaper puede detectar el job huérfano y reagendarlo. Esto implementa el patrón de consumidores competidores @hohpe2003enterprise.
- Pub/Sub + historial de eventos: el progreso se publica en un canal `Redis` y también se persiste en una lista (`{"global"}:events:{taskID}`), lo que permite a los clientes que se conectan tarde recuperar el historial completo de eventos.
- Padding en fragmentos para `blur`: al dividir la imagen para aplicar el filtro de desenfoque, cada fragmento incluye filas adicionales de `padding`, las cuales se recortan después del procesamiento para evitar artefactos en los bordes, aplicando técnicas estándar de solapamiento @rauber2013parallel.
- `SetNX` para reconstrucción única: un único worker dispara el job de reconstrucción usando `SETNX` en `Redis`, evitando que múltiples workers inicien la reconstitución simultáneamente cuando el contador llega a cero.

== Arquitectura de Componentes e Infraestructura

=== Arquitectura del Frontend
La arquitectura del cliente web, implementada con Astro y React, se visualiza en la @fig-frontend-architecture. Se compone de componentes React interactivos montados sobre páginas estáticas de Astro, integrando el hook de eventos SSE y el historial local.

#figure(
  mermaid(
    "flowchart LR

    subgraph Astro[\"Astro 6 + React 19\"]
        direction LR
        IDX[\"index.astro<br/>/\"]
        APP[\"app.astro<br/>/app · client:load\"]
        UPL[\"ImageUploader\"]
        SEL[\"FilterSelector\"]
        TRK[\"ProgressTracker\"]
        PREV[\"ResultPreview\"]
        HIST[\"TaskHistory\"]
        SSE[\"useSSE<br/>reconnect\"]
        HISTH[\"useTaskHistory<br/>localStorage\"]
        CLIENT[\"lib/api.ts<br/>POST /process · GET /result · SSE /events\"]
    end

    subgraph Stack[\"Stack\"]
        TW[\"Tailwind v4\"]
        SH[\"shadcn/ui\"]
        BIO[\"Biome\"]
        BUN[\"Bun\"]
    end

    subgraph Backend[\"Backend\"]
        API[\"Coordinator<br/>:8080\"]
    end

    IDX -.link.-> APP
    APP --> UPL
    APP --> SEL
    APP --> TRK
    APP --> PREV
    APP --> HIST
    UPL --> CLIENT
    TRK --> CLIENT
    PREV --> CLIENT
    TRK --> SSE
    HIST --> HISTH
    CLIENT -->|PUBLIC_API_URL| API

    TW -.styles.-> Astro
    SH -.components.-> Astro
    BIO -.lints.-> Astro
    BUN -.runs.-> Astro

    classDef ui fill:#1e3a8a,stroke:#3b82f6,color:#dbeafe
    classDef hook fill:#0f766e,stroke:#14b8a6,color:#ccfbf1
    classDef api fill:#7c2d12,stroke:#ea580c,color:#fed7aa
    classDef stack fill:#581c87,stroke:#a855f7,color:#f3e8ff
    classDef back fill:#14532d,stroke:#22c55e,color:#dcfce7

    class IDX,APP,UPL,SEL,TRK,PREV,HIST ui
    class SSE,HISTH hook
    class CLIENT api
    class TW,SH,BIO,BUN stack
    class API back",
  ),
  caption: [Diagrama de Componentes del Frontend de Microphoto],
) <fig-frontend-architecture>

=== Despliegue e Infraestructura Kubernetes
Para entornos de producción y escalabilidad horizontal, se ha diseñado el esquema de infraestructura Kubernetes representado en la @fig-kubernetes-architecture, incorporando balanceo de carga, TLS y monitoreo mediante Prometheus y Node Exporter.

#figure(
  mermaid(
    "flowchart LR

    subgraph Users[\"Users\"]
        U1[\"User A\"]
        U2[\"User B\"]
        UN[\"User N\"]
    end

    subgraph Edge[\"Edge\"]
        DNS[\"DNS\"]
        LB[\"LB<br/>Traefik\"]
        TLS[\"TLS\"]
    end

    subgraph Mon[\"Monitoring\"]
        PROM[\"Prometheus<br/>:9090\"]
        NEX[\"Node Exporter<br/>:9100\"]
    end

    subgraph Apps[\"Apps\"]
        COORD[\"Coordinator<br/>:8080\"]
        W1[\"Worker\"]
        W2[\"Worker\"]
        WN[\"Worker ×N\"]
        REAP[\"Reaper<br/>:9092\"]
    end

    subgraph State[\"State\"]
        REDIS[(\"Redis 7<br/>queue · pub/sub\")]
        MINIO[(\"MinIO<br/>bucket: microphoto\")]
    end

    U1 --> DNS
    U2 --> DNS
    UN --> DNS
    DNS --> LB
    LB --> TLS
    TLS -->|/api/v1/*| COORD

    COORD -->|LPUSH · PUBLISH| REDIS
    W1 -->|BLMOVE · PUBLISH| REDIS
    W2 -->|BLMOVE · PUBLISH| REDIS
    WN -->|BLMOVE · PUBLISH| REDIS
    REAP -->|SCAN · reschedule| REDIS

    COORD -->|PUT · GET| MINIO
    W1 -->|PUT · GET| MINIO
    W2 -->|PUT · GET| MINIO
    WN -->|PUT · GET| MINIO

    W1 -.queue.- W2
    W1 -.queue.- WN
    W2 -.queue.- WN

    PROM -->|scrape| COORD
    PROM -->|scrape| W1
    PROM -->|scrape| W2
    PROM -->|scrape| WN
    PROM -->|scrape| REAP
    PROM -->|scrape| NEX

    classDef user fill:#1e293b,stroke:#64748b,color:#e2e8f0
    classDef edge fill:#7c2d12,stroke:#ea580c,color:#fed7aa
    classDef mon fill:#164e63,stroke:#06b6d4,color:#cffafe
    classDef app fill:#1e3a8a,stroke:#3b82f6,color:#dbeafe
    classDef work fill:#14532d,stroke:#22c55e,color:#dcfce7
    classDef state fill:#7c2d12,stroke:#fbbf24,color:#fef3c7

    class U1,U2,UN user
    class DNS,LB,TLS edge
    class PROM,NEX mon
    class COORD,REAP app
    class W1,W2,WN work
    class REDIS,MINIO state",
  ),
  caption: [Diagrama de Despliegue en Kubernetes e Infraestructura],
) <fig-kubernetes-architecture>

#pagebreak()
= Avance actual

== Trabajo completado

=== Backend
El backend está compuesto por tres servicios principales y una base de clientes compartida:
- Coordinador: gestiona las solicitudes de procesamiento de imágenes entrantes mediante un servidor HTTP, sube las imágenes originales a MinIO, divide la carga en subtareas empujándolas a la cola Redis y ofrece retroalimentación en tiempo real a través de Server-Sent Events (SSE) @appelqvist2017performance.
- Worker: ejecuta el procesamiento paralelo consumiendo tareas de Redis. Realiza la segmentación de las imágenes en subtareas de un millón de píxeles, aplica filtros de escala de grises, brillo, desenfoque gaussiano con padding y redimensionamiento proporcional mediante bimg y libvips @h2non-bimg @libvips, y realiza la reconstrucción de la imagen resultante.
- Reaper: monitorea las tareas en proceso en Redis para detectar tiempos de expiración superados. Implementa una lógica de reintento automático y marcas de fallo definitivo cuando se agotan los intentos de procesamiento.
- Clientes y Utilidades: implementa wrappers de conexión para Redis y MinIO, incluyendo políticas de ciclo de vida para la expiración de datos temporales, y el soporte de métricas y telemetría de tareas para su consumo desde Prometheus.

=== Frontend
El frontend se estructura como una interfaz de usuario interactiva y responsiva:
- Aplicación Principal: controla el ciclo de estados de carga, procesamiento y previsualización de resultados mediante la integración con la API del coordinador.
- Panel de Control: incluye componentes para la subida de archivos mediante arrastrar y soltar, la selección y configuración detallada de parámetros de filtros, el seguimiento en tiempo real de workers activos con visualización de logs y un listado persistente con el historial local de tareas recientes.

=== Infraestructura
El entorno de despliegue abarca configuraciones locales y distribuidas:
- Entorno Local: utiliza Docker Compose para orquestar los contenedores de Redis, MinIO, coordinador, reaper y workers locales con capacidad de escalabilidad horizontal.
- Entorno Kubernetes: define manifiestos de red, ingress, gestión de certificados e integración con Helm charts y Helmfile @helmfile para gestionar despliegues de forma declarativa con secretos cifrados mediante SOPS @getsops.

#pagebreak()
= Evidencias
Se ha verificado el correcto funcionamiento del sistema en sus diferentes componentes:

- Código fuente en el repositorio: modulo Go verificado, dependencias declaradas y linter configurado.
- Artefactos de construcción: Dockerfile multi-stage con tres targets validados.
- Entorno local funcional: configuracion de Docker Compose que permite levantar el cluster local escalando el numero de workers.
- Esquema de mensajería: codigo generado a partir de la definición de Protocol Buffers mediante buf.
- Frontend compilable: compilacion estatica con bun run build.
- Capturas de la interfaz de usuario: las evidencias del funcionamiento visual de la interfaz se detallan en las figuras a continuación.

#figure(
  image("img/ev-pagina-principal.png", width: 85%),
  caption: [Página de inicio informativa (Landing page) de Microphoto],
) <fig-ev-landing>

#figure(
  image("img/ev-configuracion-antes-de-enviar-tarea.png", width: 85%),
  caption: [Selección de filtros y parámetros configurables en la aplicación],
) <fig-ev-config>

#figure(
  image("img/ev-procesamiento-de-imagen-en-progreso.png", width: 85%),
  caption: [Progreso general de las fases de procesamiento distribuido],
) <fig-ev-progress>

#figure(
  image("img/ev-procesamiento-en-progreso-con-detalle-por-nodo.png", width: 85%),
  caption: [Detalle expandido con el log individual de cada nodo worker activo],
) <fig-ev-detail>

#figure(
  image("img/ev-resultado-final.png", width: 85%),
  caption: [Previsualización de la imagen resultante procesada y opción de descarga],
) <fig-ev-result>

#pagebreak()
= Referencias
#set text(size: 11pt)
#bibliography("bibliography.bib", style: "apa", title: none)
