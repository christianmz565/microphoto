# Microphoto: Plataforma Distribuida de Procesamiento de Imágenes y Video

Microphoto es una plataforma web distribuida y escalable diseñada para el procesamiento de imágenes y video en paralelo. Implementa un pipeline de **dividir-procesar-reconstruir** que descompone archivos multimedia en fragmentos equilibrados por volumen de píxeles, distribuye las tareas en un clúster heterogéneo a través de una cola de mensajería y reensambla de forma segura el resultado final mediante sincronización distribuida.

Para mayor información técnica detallada, puede consultar el reporte del proyecto en [docs/hito2/microphoto.typ](file:///home/cricro/projects/microphoto/docs/hito2/microphoto.typ) o revisar las pautas de desarrollo en [AGENTS.md](file:///home/cricro/projects/microphoto/AGENTS.md).

---

## Arquitectura del Sistema

El sistema sigue una arquitectura productor-cola-consumidor con almacenamiento compartido de objetos y consta de cinco componentes clave:

1. **Coordinator** ([backend/cmd/coordinator](file:///home/cricro/projects/microphoto/backend/cmd/coordinator)): API HTTP en Go que expone endpoints para subida de medios, orquestación de tareas y streaming de eventos a través de Server-Sent Events (SSE) en tiempo real.
2. **Worker** ([backend/cmd/worker](file:///home/cricro/projects/microphoto/backend/cmd/worker)): Consumidor de tareas en Go que se encarga del procesamiento de fragmentos de imagen/video (escala de grises, desenfoque gaussiano, brillo, redimensionamiento) en paralelo mediante libvips (vía `bimg`).
3. **Reaper** ([backend/cmd/reaper](file:///home/cricro/projects/microphoto/backend/cmd/reaper)): Servicio de tolerancia a fallos en Go que monitorea las tareas activas y reprograma aquellos trabajos que exceden el tiempo límite de procesamiento (timeout).
4. **Redis**: Cola FIFO global y almacén de estado distribuido. Utiliza operaciones atómicas (`BLMOVE`, `DECR`, `SetNX`) para la gestión fiable de tareas, sincronización de barrera y señalización pub/sub de progreso.
5. **MinIO / Garage**: Almacenamiento de objetos compatible con S3 para guardar los archivos multimedia originales, fragmentos intermedios y el resultado reconstruido.

---

## Flujo de Datos

1. **Subida**: El usuario envía una imagen o video al Coordinator, el cual la almacena en MinIO/Garage y encola una tarea de fragmentación (`SLICE`).
2. **División**: Un Worker extrae el trabajo `SLICE`, fragmenta la imagen en $N$ partes proporcionales (apuntando a 1 megapíxel por fragmento) y las sube a S3, encolando seguidamente las $N$ tareas de procesamiento.
3. **Procesamiento**: Los Workers procesan los fragmentos de manera distribuida. Cada vez que un fragmento finaliza, el Worker decrementa atómicamente un contador de tareas pendientes en Redis y notifica el progreso.
4. **Reconstrucción**: El Worker que reduce el contador a 0 adquiere un bloqueo distribuido (`SetNX`) y procede a reensamblar la imagen final combinando los fragmentos procesados (aplicando padding para evitar artefactos en los bordes).
5. **Video (Descomposición Temporal)**: Para los videos, se divide temporalmente en segmentos de 3 segundos usando `ffmpeg`, se procesan sus fotogramas y finalmente se concatenan los segmentos reconstruidos.

---

## Requisitos de Entorno

* **Nix / direnv** (Recomendado): Proporciona todo el toolchain configurado automáticamente.
* **Go**: Versión 1.21+ (con CGO habilitado únicamente para la compilación del Worker).
* **Node.js**: Versión >= 22.12.0.
* **Bun**: Gestor de dependencias y motor de ejecución frontend.
* **libvips**: Requerido por el motor de procesamiento del Worker.
* **ffmpeg**: Requerido para la fragmentación y ensamblaje de video.
* **Docker y Docker Compose**: Para levantar los servicios de infraestructura de manera local.

---

## Instrucciones de Uso

### 1. Levantar la Infraestructura Local

Inicia los servicios de base (Redis, MinIO, Prometheus, Grafana) y los contenedores de la aplicación:

```bash
docker compose up
```

La configuración por defecto se encuentra definida en [docker-compose.yml](file:///home/cricro/projects/microphoto/docker-compose.yml).

### 2. Compilación del Backend (`backend/`)

Para compilar localmente los binarios de Go, se requiere la utilidad `just`. Ejecute los siguientes comandos en la raíz del proyecto o dentro del directorio correspondiente:

```bash
just build                  # Compila coordinator, worker y reaper
just docker-all             # Construye las imágenes Docker para todos los servicios
just proto                  # Regenera el código Go de Protobuf
golangci-lint run           # Ejecuta el análisis estático de código (linter)
```

> [!IMPORTANT]
> La compilación del Worker requiere `CGO_ENABLED=1` y la presencia de la librería libvips en el sistema. El Coordinator y el Reaper pueden ser compilados con CGO deshabilitado.

### 3. Ejecución del Frontend (`frontend/`)

El frontend está desarrollado con Astro 7 e islas de React. Inicie el servidor de desarrollo local:

```bash
cd frontend
bun install                 # Instala dependencias
bun dev                     # Inicia el servidor de desarrollo en http://localhost:4321
bun run build               # Genera el build de producción en ./dist
bunx biome check .          # Ejecuta verificación de formato y linter (Biome)
bunx biome check --write .  # Aplica correcciones automáticas de formato
```

---

## Monitoreo y Observabilidad

El sistema expone métricas nativas de OpenTelemetry en los puertos `9090` (Coordinator), `9091` (Worker) y `9092` (Reaper). 

* **Prometheus**: Recolecta las métricas de rendimiento y timeouts de las tareas.
* **Grafana**: Permite visualizar a través de paneles preconfigurados la utilización de CPU y memoria de los nodos del clúster, tiempos de procesamiento de tareas y tasa de fallos. El aprovisionamiento se gestiona en la carpeta [monitoring](file:///home/cricro/projects/microphoto/monitoring).
