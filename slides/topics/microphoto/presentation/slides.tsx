import { Slide } from "@revealjs/react";
import { Badge } from "@/shared/badge";
import { colors as C } from "@/shared/colors";
import {
  FeatureCard,
  FeatureCardCompact,
  FeatureCardCompactSmall,
  FeatureCardTall,
} from "@/shared/feature-cards";
import { NumberedItem } from "@/shared/numbered-item";
import { PresentationDeck } from "@/shared/presentation-deck";
import { SlideWrap } from "@/shared/slide-wrap";
import { StatCard } from "@/shared/stat-cards";
import { ThanksSlide } from "@/shared/thanks-slide";

const baseUrl = import.meta.env.BASE_URL.replace(/\/$/, "");

const LOGO = `${baseUrl}/microphoto/logo.png`;
const IMG_SEQ_1 = `${baseUrl}/microphoto/diagrama-secuencia-parte-1.png`;
const IMG_SEQ_2 = `${baseUrl}/microphoto/diagrama-secuencia-parte-2.png`;
const IMG_EV_LANDING = `${baseUrl}/microphoto/ev-pagina-principal.png`;
const IMG_ARCHITECTURE = `${baseUrl}/microphoto/diagrama-arquitectura.png`;

/* ─── PLACEHOLDER PARA CAPTURAS PENDIENTES ─── */
function ImagePlaceholder({ label }: { label: string }) {
  return (
    <div
      className="flex flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed p-8 text-center"
      style={{ borderColor: `${C.petrol}50`, background: `${C.petrol}0a` }}
    >
      <svg
        role="img"
        aria-label="Espacio para captura de pantalla"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.petrol}
        strokeWidth="1.5"
        className="size-10 opacity-70"
      >
        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
        <circle cx="8.5" cy="8.5" r="1.5" />
        <polyline points="21 15 16 10 5 21" />
      </svg>
      <p
        className="text-lg font-semibold uppercase tracking-[0.15em]"
        style={{ color: C.petrol }}
      >
        Espacio para captura
      </p>
      <p className="text-lg text-gray-400 max-w-md text-pretty">{label}</p>
    </div>
  );
}

/* ─── TARJETA DE REQUERIMIENTO ─── */
function statusColor(status: "Completo" | "En curso" | "Pendiente") {
  if (status === "Completo") return C.green;
  if (status === "En curso") return C.amber;
  return C.red;
}

function RequirementCard({
  code,
  desc,
  status,
}: {
  code: string;
  desc: string;
  status: "Completo" | "En curso" | "Pendiente";
}) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-3 flex flex-col gap-1.5 text-left">
      <div className="flex items-center justify-between">
        <Badge label={code} color={C.petrol} />
        <span
          className="text-sm font-semibold uppercase tracking-wider"
          style={{ color: statusColor(status) }}
        >
          {status}
        </span>
      </div>
      <p className="text-lg text-gray-300 leading-snug">{desc}</p>
    </div>
  );
}

/* ─── PORTADA + INTEGRANTES ─── */
function Cover() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Sistemas Distribuidos · Grupo 2"
        variant="decorated"
        className="justify-center flex flex-col items-center h-full py-8 text-center"
      >
        <img src={LOGO} alt="Microphoto" className="h-28 object-contain" />
        <h1 className="mt-2 text-7xl!">
          <span style={{ color: C.petrol }}>Microphoto</span>
        </h1>
        <p className="mt-2 text-2xl text-gray-400 max-w-3xl text-pretty">
          Procesamiento paralelo de imágenes y video sobre una arquitectura de
          sistemas distribuidos
        </p>
        <div className="mt-6">
          <p className="text-xl font-semibold uppercase tracking-[0.2em] text-gray-500">
            Integrantes
          </p>
          <p className="mt-1.5 text-2xl text-gray-300 font-light max-w-4xl text-pretty">
            Bedregal Pérez, Daniel · Jara Mamani, Mariel Alisson · Mestas
            Zegarra, Christian Raúl · Noa Camino, Yenaro Joel · Sequeiros
            Condori, Luis Gustavo
          </p>
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── MOTIVACIÓN ─── */
function Motivation() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="¿Por qué?"
        variant="decorated"
        className="justify-center flex flex-col gap-6 w-full mx-auto"
      >
        <div className="flex flex-col items-center">
          <h1 className="text-5xl! text-balance max-w-4xl text-center">
            Del procesamiento secuencial al{" "}
            <span style={{ color: C.petrol }}>procesamiento distribuido</span>
          </h1>
          <p className="mt-2 text-xl text-gray-400 max-w-3xl text-center text-pretty">
            Microphoto fragmenta, procesa en paralelo y reconstruye imágenes y
            video de gran tamaño, distribuyendo el trabajo entre múltiples nodos
            worker.
          </p>
        </div>

        <div className="grid grid-cols-3 gap-3">
          {[
            {
              l: "Pipeline distribuido",
              d: "Fragmentación automática, distribución entre workers y reconstrucción del resultado final.",
            },
            {
              l: "Feedback en tiempo real",
              d: "Retroalimentación continua del estado del procesamiento mediante Server-Sent Events.",
            },
            {
              l: "Escalabilidad horizontal",
              d: "Réplicas de workers en Docker Compose o Kubernetes sin modificar el código.",
            },
          ].map((i) => (
            <FeatureCard
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
              variant="decorated"
            />
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-center gap-3">
          <span className="text-xl font-semibold uppercase tracking-[0.15em] text-gray-500">
            Capacidades actuales
          </span>
          {[
            "Imágenes",
            "Video",
            "Vista previa en vivo",
            "Efectos encadenados",
          ].map((f) => (
            <Badge key={f} label={f} color={C.petrol} />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── REQUERIMIENTOS: DIVISOR ─── */
function RequirementsDivider() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Requerimientos"
        variant="decorated"
        className="justify-center items-center flex flex-col gap-4"
      >
        <h1 className="text-6xl! font-semibold tracking-tight text-center">
          Funcionales <span className="text-gray-500 font-light">y</span>{" "}
          <span style={{ color: C.petrol }}>No Funcionales</span>
        </h1>
        <p className="text-xl text-gray-400 max-w-2xl text-center text-pretty">
          Verificados directamente sobre el estado actual del código del backend
          y el frontend.
        </p>
      </SlideWrap>
    </Slide>
  );
}

/* ─── REQUERIMIENTOS FUNCIONALES ─── */
function FunctionalRequirements() {
  const items: {
    code: string;
    desc: string;
    status: "Completo" | "En curso" | "Pendiente";
  }[] = [
    {
      code: "RF-001",
      desc: "Subida de imágenes y video desde el navegador (hasta 2 GB).",
      status: "Completo",
    },
    {
      code: "RF-002",
      desc: "Fragmentación por píxeles (imagen) o segmentación por tiempo (video).",
      status: "Completo",
    },
    {
      code: "RF-003",
      desc: "Procesamiento paralelo en workers con efectos encadenables.",
      status: "Completo",
    },
    {
      code: "RF-004",
      desc: "Vista previa en vivo de los efectos sin pasar por el clúster.",
      status: "Completo",
    },
    {
      code: "RF-005",
      desc: "Reconstrucción y reensamblado automático del resultado final.",
      status: "Completo",
    },
    {
      code: "RF-006",
      desc: "Progreso en tiempo real vía SSE, con detalle por worker.",
      status: "Completo",
    },
    {
      code: "RF-007",
      desc: "Historial local de tareas recientes en el navegador.",
      status: "Completo",
    },
    {
      code: "RF-008",
      desc: "Historial de tareas persistente en base de datos (hoy solo en localStorage).",
      status: "Completo",
    },
  ];
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Requerimientos Funcionales"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-4xl! font-semibold tracking-tight text-center">
          Qué puede hacer <span style={{ color: C.petrol }}>hoy</span> la
          plataforma
        </h1>
        <div className="grid grid-cols-3 gap-2.5">
          {items.map((i) => (
            <RequirementCard key={i.code} {...i} />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── REQUERIMIENTOS NO FUNCIONALES ─── */
function NonFunctionalRequirements() {
  const items: {
    code: string;
    desc: string;
    status: "Completo" | "En curso" | "Pendiente";
  }[] = [
    {
      code: "RNF-001",
      desc: "Escalabilidad horizontal: workers replicables sin cambiar código.",
      status: "Completo",
    },
    {
      code: "RNF-002",
      desc: "Tolerancia a fallos: reaper con reintentos (máx. 3) y colas atómicas.",
      status: "Completo",
    },
    {
      code: "RNF-003",
      desc: "Contenerización multi-stage con Docker para los tres servicios.",
      status: "Completo",
    },
    {
      code: "RNF-004",
      desc: "Vista previa en menos de 2 segundos de respuesta.",
      status: "Completo",
    },
    {
      code: "RNF-005",
      desc: "Observabilidad: métricas OpenTelemetry/Prometheus por servicio.",
      status: "Completo",
    },
    {
      code: "RNF-006",
      desc: "Despliegue en Kubernetes con Helm charts y Helmfile.",
      status: "Completo",
    },
    {
      code: "RNF-007",
      desc: "Gestión de secretos cifrados con SOPS.",
      status: "Completo",
    },
  ];
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Requerimientos No Funcionales"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-4xl! font-semibold tracking-tight text-center">
          Cómo se comporta{" "}
          <span style={{ color: C.petrol }}>bajo carga y fallos</span>
        </h1>
        <div className="grid grid-cols-4 gap-2.5">
          {items.map((i) => (
            <RequirementCard key={i.code} {...i} />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── ARQUITECTURA GLOBAL ─── */
function Architecture() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Arquitectura · Visión Global"
        variant="decorated"
        className="w-full flex flex-col gap-2 items-center justify-center"
      >
        <div className="flex flex-col items-center">
          <h1 className="text-5xl! text-balance text-center">
            Patrón{" "}
            <span className="font-semibold" style={{ color: C.petrol }}>
              productor-cola-consumidor
            </span>
          </h1>
          <p className="mt-1 text-xl text-gray-400 text-pretty text-center max-w-4xl">
            Coordinador, cola Redis, workers y almacenamiento MinIO orquestados
            detrás de un balanceador Traefik, con métricas expuestas a
            Prometheus.
          </p>
        </div>
        <div className="grid grid-cols-2 gap-2 self-stretch">
          {[
            {
              l: "Cola fiable BLMOVE",
              d: "Movimiento atómico de jobs; el reaper reagenda los huérfanos.",
            },
            {
              l: "Pub/Sub + historial",
              d: "Progreso publicado y persistido para clientes que se conectan tarde.",
            },
            {
              l: "Padding en blur",
              d: "Filas adicionales por fragmento para evitar artefactos en bordes.",
            },
            {
              l: "SetNX de reconstrucción",
              d: "Un único worker dispara la reconstitución final del resultado.",
            },
          ].map((i) => (
            <FeatureCardCompactSmall
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
            />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

function ImgArchitecture() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Arquitectura · Visión Global"
        variant="decorated"
        className="w-full h-full flex items-center justify-center"
      >
        <div className="flex flex-col items-center">
          <img
            src={IMG_ARCHITECTURE}
            alt="Diagrama de arquitectura global: coordinador, workers, reaper, Redis, MinIO y Traefik"
            className="object-contain rounded-xl bg-white/95 p-2 shadow-md w-md"
          />
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── CICLO DE VIDA: SECUENCIA (IMAGEN) ─── */
function SequencePart1() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Ciclo de Vida (Imagen) · Secuencia 1/2"
        variant="decorated"
        className="w-full flex items-center justify-center"
      >
        <div className="grid grid-cols-2 gap-4 w-full items-center text-left px-2">
          <div className="flex items-center justify-center p-2">
            <img
              src={IMG_SEQ_1}
              alt="Diagrama de secuencia: subida, suscripción y corte"
              className="object-contain rounded-xl bg-white/95 p-2 shadow-md w-xl"
            />
          </div>
          <div className="flex flex-col gap-4">
            <h1 className="text-3xl! font-semibold tracking-tight">
              Subida,{" "}
              <span style={{ color: C.petrol }}>suscripción y corte</span>
            </h1>
            <NumberedItem
              num="1"
              title="Subida y creación de tarea"
              description="El usuario envía la imagen al coordinador, que la guarda en MinIO y encola el trabajo de corte en Redis."
              color={C.petrol}
            />
            <NumberedItem
              num="2"
              title="Suscripción a progreso"
              description="El coordinador se suscribe al canal de progreso de Redis asociado a la tarea."
              color={C.petrol}
            />
            <NumberedItem
              num="3"
              title="Corte de imagen"
              description="Un worker fragmenta la imagen, guarda las partes en MinIO, inicializa contadores y encola los trabajos de procesamiento."
              color={C.petrol}
            />
          </div>
        </div>
      </SlideWrap>
    </Slide>
  );
}

function SequencePart2() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Ciclo de Vida (Imagen) · Secuencia 2/2"
        variant="decorated"
        className="w-full flex items-center justify-center"
      >
        <div className="grid grid-cols-2 gap-4 w-full items-center text-left">
          <div className="flex flex-col gap-4">
            <h1 className="text-3xl! font-semibold tracking-tight">
              Procesamiento,{" "}
              <span style={{ color: C.petrol }}>reconstrucción y entrega</span>
            </h1>
            <NumberedItem
              num="4"
              title="Procesamiento paralelo"
              description="Cada worker toma un fragmento, aplica los efectos configurados y reporta su progreso, que se emite al coordinador."
              color={C.petrol}
            />
            <NumberedItem
              num="5"
              title="Reconstrucción"
              description="Un único worker, mediante SETNX, toma el cerrojo de reconstrucción, compone la imagen final y notifica su finalización."
              color={C.petrol}
            />
            <NumberedItem
              num="6"
              title="Descarga y visualización"
              description="El usuario solicita el resultado; el coordinador lo obtiene de MinIO y lo entrega al navegador."
              color={C.petrol}
            />
          </div>
          <div className="flex items-center justify-center p-2">
            <img
              src={IMG_SEQ_2}
              alt="Diagrama de secuencia: procesamiento, reconstrucción y descarga"
              className="object-contain rounded-xl bg-white/95 p-2 shadow-md w-xl"
            />
          </div>
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── PIPELINE DE VIDEO (NUEVO) ─── */
function VideoPipeline() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Pipeline de Video · Novedad"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center items-center"
      >
        <div className="flex flex-col items-center">
          <h1 className="text-5xl! text-balance text-center">
            Dos niveles de <span style={{ color: C.petrol }}>paralelismo</span>
          </h1>
          <p className="mt-1 text-xl text-gray-400 text-pretty text-center max-w-4xl">
            El video no se corta por píxeles como la imagen: se segmenta por
            tiempo y luego se paraleliza por frame dentro de cada segmento.
          </p>
        </div>
        <div className="grid grid-cols-4 gap-2.5 self-stretch">
          {[
            {
              l: "Segmentación temporal",
              d: "ffmpeg divide el video en segmentos de 3s por defecto, subidos y encolados en paralelo.",
            },
            {
              l: "Extracción de frames",
              d: "Cada segmento se descompone en frames JPEG (ffmpeg -q:v 2).",
            },
            {
              l: "Filtrado paralelo",
              d: "El mismo pipeline de efectos de imagen se aplica a cada frame, con concurrencia configurable (8 por defecto).",
            },
            {
              l: "Reensamblado",
              d: "Los frames se recomponen en MP4 (libx264) y los segmentos se concatenan en el resultado final.",
            },
          ].map((i) => (
            <FeatureCardCompact
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
              variant="decorated"
            />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── BACKEND ─── */
function Backend() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Componentes · Backend"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Tres servicios en <span style={{ color: C.petrol }}>Go</span>
        </h1>
        <div className="grid grid-cols-3 gap-3">
          {[
            {
              l: "Coordinador",
              d: "Recibe subidas de imagen y video, expone la vista previa en vivo y transmite el progreso vía SSE.",
            },
            {
              l: "Worker",
              d: "Consume tareas de Redis; fragmenta imágenes por píxeles o segmenta video por tiempo, y aplica efectos encadenados con bimg/libvips y ffmpeg.",
            },
            {
              l: "Reaper",
              d: "Detecta tareas colgadas cada 5s; reagenda hasta 3 intentos por tarea o marca fallo definitivo.",
            },
          ].map((i) => (
            <FeatureCardCompact
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
              variant="decorated"
            />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── FRONTEND: EDITOR DE EFECTOS ─── */
function Frontend() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Componentes · Frontend"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Editor con <span style={{ color: C.petrol }}>efectos en vivo</span>
        </h1>
        <div className="grid grid-cols-2 gap-3">
          <FeatureCardCompact
            label="Editor de Efectos"
            description="Sliders de escala de grises, desenfoque y brillo con vista previa en vivo; permite descarga rápida en el cliente o procesamiento distribuido en el clúster."
            color={C.petrol}
            variant="decorated"
          />
          <FeatureCardCompact
            label="Panel de Control"
            description="Subida por arrastrar y soltar, selección de archivo o pegado (Ctrl+V) de imagen o video, y un historial local de tareas recientes."
            color={C.petrol}
            variant="decorated"
          />
        </div>
        <div className="flex flex-wrap items-center justify-center gap-2">
          {["Grises", "Desenfoque", "Brillo", "Contraste", "Sepia", "Viñeta", "Escala"].map((s) => (
            <Badge key={s} label={s} color={C.petrol} />
          ))}
        </div>
        <div className="flex items-center justify-center gap-3">
          <span className="text-xl font-semibold uppercase tracking-[0.15em] text-gray-500">
            Stack
          </span>
          {[
            "Astro",
            "React 19",
            "Tailwind v4",
            "shadcn/ui",
            "Bun",
            "Biome",
          ].map((s) => (
            <Badge key={s} label={s} color={C.petrol} />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── INFRAESTRUCTURA (KUBERNETES) ─── */
function Infrastructure() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Infraestructura · Kubernetes en producción"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Un clúster{" "}
          <span style={{ color: C.petrol }}>completo y desplegado</span>
        </h1>
        <div className="grid grid-cols-3 gap-2.5">
          {[
            {
              l: "Entornos dev y prod",
              d: "Namespaces separados, manifiestos gestionados con Kustomize y overlays por entorno.",
            },
            {
              l: "Ingress y TLS",
              d: "Traefik enruta el tráfico; cert-manager emite certificados Let's Encrypt vía DNS-01 con Cloudflare.",
            },
            {
              l: "DNS automatizado",
              d: "external-dns sincroniza los registros del dominio con el estado real del clúster.",
            },
            {
              l: "Secretos cifrados",
              d: "SOPS + age (5 llaves) protege los secretos de cada servicio y de ambos entornos.",
            },
            {
              l: "Observabilidad",
              d: "Prometheus, Grafana y Node Exporter recolectan y visualizan métricas del clúster.",
            },
            {
              l: "Autoescalado horizontal",
              d: "Un HorizontalPodAutoscaler ajusta las réplicas del worker según el uso de CPU.",
            },
          ].map((i) => (
            <FeatureCardCompactSmall
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
            />
          ))}
        </div>
        <div className="grid grid-cols-3 gap-2.5">
          <StatCard
            value="2 – 20"
            label="Réplicas del worker (HPA)"
            color={C.petrol}
            variant="decorated"
          />
          <StatCard
            value="70%"
            label="CPU objetivo por réplica"
            color={C.petrol}
            variant="decorated"
          />
          <StatCard
            value="5"
            label="Llaves age para SOPS"
            color={C.petrol}
            variant="decorated"
          />
        </div>
      </SlideWrap>
    </Slide>
  );
}

function Evidence({
  index,
  total,
  title,
  highlight,
  desc,
  image,
  placeholder,
}: {
  index: number;
  total: number;
  title: string;
  highlight: string;
  desc: string;
  image?: string;
  placeholder?: string;
}) {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag={`Evidencias · ${index}/${total}`}
        variant="decorated"
        className="w-full flex items-center justify-center"
      >
        <div className="grid grid-cols-2 gap-8 w-full items-center text-left px-8">
          <div className="flex flex-col gap-4">
            <h1 className="text-4xl! font-semibold tracking-tight leading-tight">
              {title} <span style={{ color: C.petrol }}>{highlight}</span>
            </h1>
            <p className="text-xl text-gray-400 text-pretty">{desc}</p>
          </div>
          <div className="flex items-center justify-center">
            {image ? (
              <img
                src={image}
                alt={`${title} ${highlight}`}
                className="object-contain rounded-xl bg-white/95 p-2 shadow-md max-h-[440px] w-full"
              />
            ) : (
              <ImagePlaceholder label={placeholder ?? ""} />
            )}
          </div>
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── TRABAJO FUTURO ─── */
function PendingWork() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Trabajo Futuro"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center items-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Camino al <span style={{ color: C.petrol }}>cierre del proyecto</span>
        </h1>
        <div className="grid grid-cols-2 gap-3 self-stretch">
          {[
            {
              l: "Autenticación de usuarios",
              d: "Registro e inicio de sesión seguro para proteger datos y personalizar el acceso.",
            },
            {
              l: "Autoescalado por profundidad de cola",
              d: "Extender el HPA actual (basado en CPU) para reaccionar al tamaño real de la cola en Redis, escalando workers según el backlog de tareas y no solo el uso de CPU.",
            },
            {
              l: "Procesamiento por lotes",
              d: "Permitir subir múltiples imágenes o videos en una sola tarea, procesándolos en paralelo y entregando los resultados como un paquete comprimido.",
            },
            {
              l: "Reensamblado de video con GPU",
              d: "Acelerar la fase REASSEMBLING con codificación por hardware (ffmpeg + NVENC) para acortar el tiempo de entrega en clips largos.",
            },
          ].map((i) => (
            <FeatureCardTall
              key={i.l}
              label={i.l}
              description={i.d}
              color={C.petrol}
            />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── CONCLUSIONES ─── */
function Conclusions() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Conclusiones"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-6xl! font-semibold tracking-tight text-center">
          Un pipeline{" "}
          <span style={{ color: C.petrol }}>distribuido y funcional</span>
        </h1>
        <div className="grid grid-cols-3 gap-4">
          {[
            {
              n: "1",
              l: "Arquitectura validada y extendida",
              d: "El patrón productor-cola-consumidor con Redis y MinIO ahora soporta imagen y video sobre la misma base.",
            },
            {
              n: "2",
              l: "Escalabilidad horizontal",
              d: "Los workers se replican sin cambios de código, tanto en Docker Compose como en Kubernetes.",
            },
            {
              n: "3",
              l: "Tolerancia a fallos",
              d: "El reaper y las colas atómicas (BLMOVE, SETNX) evitan que una tarea se pierda ante una caída.",
            },
          ].map((i) => (
            <div
              key={i.n}
              className="rounded-xl border border-white/10 bg-white/5 p-2"
            >
              <div className="flex items-center gap-2">
                <span
                  className="flex size-5 shrink-0 items-center justify-center rounded-full text-base font-semibold"
                  style={{ backgroundColor: `${C.petrol}25`, color: C.petrol }}
                >
                  {i.n}
                </span>
                <p className="text-xl font-semibold uppercase tracking-widest text-white">
                  {i.l}
                </p>
              </div>
              <p className="mt-0.5 text-xl text-left text-gray-400 leading-snug">
                {i.d}
              </p>
            </div>
          ))}
        </div>
        <p className="text-center text-xl text-gray-400">
          Según el backlog reportado, Microphoto avanza al{" "}
          <span className="font-semibold text-white">100 %</span>; el código ya
          incorpora capacidades no documentadas —
          <span className="font-semibold text-white">
            procesamiento de video y vista previa en vivo
          </span>
          — que amplían ese alcance.
        </p>
      </SlideWrap>
    </Slide>
  );
}

/* ─── DECK ─── */
export function MicrophotoPresentation() {
  return (
    <PresentationDeck config={{ slideNumber: "c/t", transition: "slide" }}>
      <Cover />
      <Motivation />
      <VideoPipeline />

      <RequirementsDivider />
      <FunctionalRequirements />
      <NonFunctionalRequirements />

      <Architecture />
      <ImgArchitecture />
      <SequencePart1 />
      <SequencePart2 />

      <Backend />
      <Frontend />

      <Evidence
        index={1}
        total={1}
        title="Página de inicio"
        highlight="informativa"
        desc="Landing page que explica el problema, el flujo del pipeline y las capacidades de la plataforma."
        image={IMG_EV_LANDING}
      />

      <PendingWork />
      <Conclusions />
      <ThanksSlide color={C.petrol} variant="decorated" />
    </PresentationDeck>
  );
}
