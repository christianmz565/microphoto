import { Slide } from "@revealjs/react";
import { Badge } from "@/shared/badge";
import { colors as C } from "@/shared/colors";
import { DotList } from "@/shared/dot-list";
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
const IMG_ARCH = `${baseUrl}/microphoto/arquitectura-general.png`;
const IMG_SEQ_1 = `${baseUrl}/microphoto/diagrama-secuencia-parte-1.png`;
const IMG_SEQ_2 = `${baseUrl}/microphoto/diagrama-secuencia-parte-2.png`;
const IMG_EV_LANDING = `${baseUrl}/microphoto/ev-pagina-principal.png`;
const IMG_EV_CONFIG = `${baseUrl}/microphoto/ev-configuracion-antes-de-enviar-tarea.png`;
const IMG_EV_PROGRESS = `${baseUrl}/microphoto/ev-procesamiento-de-imagen-en-progreso.png`;
const IMG_EV_DETAIL = `${baseUrl}/microphoto/ev-procesamiento-en-progreso-con-detalle-por-nodo.png`;
const IMG_EV_RESULT = `${baseUrl}/microphoto/ev-resultado-final.png`;

/* ─── PORTADA ─── */
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
          Procesamiento paralelo de imágenes sobre una arquitectura de
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

/* ─── INTRODUCCIÓN ─── */
function Introduction() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Introducción"
        variant="decorated"
        className="justify-center flex flex-col gap-6 w-full mx-auto"
      >
        <div className="flex flex-col items-center">
          <h1 className="text-5xl! text-balance max-w-4xl text-center">
            Del procesamiento secuencial al{" "}
            <span style={{ color: C.petrol }}>procesamiento distribuido</span>
          </h1>
          <p className="mt-2 text-xl text-gray-400 max-w-3xl text-center text-pretty">
            Microphoto fragmenta, procesa en paralelo y reconstruye imágenes
            de gran tamaño, distribuyendo el trabajo entre múltiples nodos
            worker.
          </p>
        </div>

        <div className="grid grid-cols-3 gap-3">
          {[
            {
              l: "Pipeline distribuido",
              d: "Fragmentación automática, distribución entre workers y reconstrucción de la imagen final.",
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

        <div className="flex items-center justify-center gap-3">
          <span className="text-xl font-semibold uppercase tracking-[0.15em] text-gray-500">
            Filtros disponibles
          </span>
          {["GRAYSCALE", "BLUR", "BRIGHTNESS", "RESIZE"].map((f) => (
            <Badge key={f} label={f} color={C.petrol} />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── EQUIPO Y ROLES ─── */
function Team() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Equipo y Roles"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center items-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight">
          Organización del{" "}
          <span style={{ color: C.petrol }}>equipo de trabajo</span>
        </h1>
        <div className="grid grid-cols-3 gap-3 self-stretch">
          {[
            {
              l: "Backend",
              d: "Bedregal Pérez, Daniel · Mestas Zegarra, Christian Raúl · Noa Camino, Yenaro Joel — coordinador HTTP, procesamiento con bimg/libvips, slicing, reconstitución y reaper.",
            },
            {
              l: "Frontend",
              d: "Jara Mamani, Mariel Alisson — aplicación con Astro y React, integración SSE y componentes UI con shadcn/ui.",
            },
            {
              l: "Infraestructura",
              d: "Sequeiros Condori, Luis Gustavo — Docker Compose, manifiestos Kubernetes, Helm charts, Helmfile y secretos con SOPS.",
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

/* ─── ESTADO GENERAL ─── */
function ProjectStatus() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Estado General del Proyecto"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Avance al{" "}
          <span style={{ color: C.petrol }}>24 de junio de 2026</span>
        </h1>
        <div className="grid grid-cols-4 gap-3">
          {[
            { v: "50%", l: "Avance general del proyecto" },
            { v: "3/3", l: "Servicios backend operativos" },
            { v: "5/5", l: "Vistas frontend con SSE" },
            { v: "13/15", l: "Requerimientos funcionales completos" },
          ].map((i) => (
            <StatCard
              key={i.l}
              value={i.v}
              label={i.l}
              color={C.petrol}
              variant="decorated"
            />
          ))}
        </div>
        <div className="rounded-xl border border-white/10 bg-white/5 p-4">
          <p className="text-lg font-semibold uppercase tracking-[0.15em] text-gray-500 mb-1">
            Pendiente para el cierre
          </p>
          <DotList
            items={[
              "Autenticación y gestión de usuarios",
              "Historial persistente de tareas en base de datos",
              "Pruebas automatizadas unitarias e integración",
              "Validación del despliegue en un clúster Kubernetes real",
            ]}
            color={C.petrol}
          />
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── BACKLOG: ÉPICAS ─── */
function Backlog() {
  const epics = [
    {
      code: "E-001",
      l: "Pipeline de procesamiento distribuido",
      v: "100%",
      c: C.green,
    },
    {
      code: "E-002",
      l: "Interfaz de usuario y experiencia de cliente",
      v: "100%",
      c: C.green,
    },
    {
      code: "E-003",
      l: "Observabilidad y monitoreo",
      v: "33%",
      c: C.amber,
    },
    {
      code: "E-004",
      l: "Infraestructura y despliegue",
      v: "50%",
      c: C.amber,
    },
    {
      code: "E-005",
      l: "Gestión de usuarios y seguridad",
      v: "0%",
      c: C.red,
    },
  ];
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Backlog · Resumen de Épicas"
        variant="decorated"
        className="w-full flex flex-col gap-4 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Avance por{" "}
          <span style={{ color: C.petrol }}>épica</span>
        </h1>
        <div className="grid grid-cols-5 gap-3">
          {epics.map((e) => (
            <div
              key={e.code}
              className="rounded-xl border border-white/10 bg-white/5 p-3 flex flex-col items-center text-center"
            >
              <Badge label={e.code} color={C.petrol} />
              <p className="mt-3 text-6xl font-bold" style={{ color: e.c }}>
                {e.v}
              </p>
              <p className="mt-2 text-lg text-gray-400 leading-snug">
                {e.l}
              </p>
            </div>
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
        <img
          src={IMG_ARCH}
          alt="Arquitectura global de Microphoto"
          className="object-contain rounded-xl bg-white/95 p-3 shadow-md max-h-72"
        />
        <div className="grid grid-cols-4 gap-2 self-stretch">
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
              d: "Un único worker dispara la reconstitución final de la imagen.",
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

/* ─── CICLO DE VIDA: SECUENCIA ─── */
function SequencePart1() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Ciclo de Vida · Secuencia 1/2"
        variant="decorated"
        className="w-full flex items-center justify-center"
      >
        <div className="grid grid-cols-2 gap-4 w-full items-center text-left">
          <div className="flex items-center justify-center p-2">
            <img
              src={IMG_SEQ_1}
              alt="Diagrama de secuencia: subida, suscripción y corte"
              className="object-contain rounded-xl bg-white/95 p-2 shadow-md max-h-[560px]"
            />
          </div>
          <div className="flex flex-col gap-2">
            <h1 className="text-4xl! font-semibold tracking-tight">
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
        tag="Ciclo de Vida · Secuencia 2/2"
        variant="decorated"
        className="w-full flex items-center justify-center"
      >
        <div className="grid grid-cols-2 gap-4 w-full items-center text-left">
          <div className="flex flex-col gap-2">
            <h1 className="text-4xl! font-semibold tracking-tight">
              Procesamiento,{" "}
              <span style={{ color: C.petrol }}>reconstrucción y entrega</span>
            </h1>
            <NumberedItem
              num="4"
              title="Procesamiento paralelo"
              description="Cada worker toma un fragmento, aplica el filtro configurado y reporta su progreso, que se emite al coordinador."
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
              className="object-contain rounded-xl bg-white/95 p-2 shadow-md max-h-[560px]"
            />
          </div>
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
          Tres servicios en{" "}
          <span style={{ color: C.petrol }}>Go</span>
        </h1>
        <div className="grid grid-cols-4 gap-3">
          {[
            {
              l: "Coordinador",
              d: "Recibe las solicitudes HTTP, sube la imagen original a MinIO, encola las subtareas y transmite progreso vía SSE.",
            },
            {
              l: "Worker",
              d: "Consume tareas de Redis, segmenta en fragmentos de un millón de píxeles y aplica filtros con bimg y libvips.",
            },
            {
              l: "Reaper",
              d: "Monitorea tareas en proceso, detecta expiraciones y reintenta o marca fallo definitivo tras agotar los intentos.",
            },
            {
              l: "Clientes y Utilidades",
              d: "Wrappers de conexión a Redis y MinIO, políticas de expiración y telemetría para Prometheus.",
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

/* ─── FRONTEND ─── */
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
          Cliente web con{" "}
          <span style={{ color: C.petrol }}>Astro y React</span>
        </h1>
        <div className="grid grid-cols-2 gap-3">
          <FeatureCardCompact
            label="Aplicación Principal"
            description="Controla el ciclo de carga, procesamiento y previsualización de resultados integrado con la API del coordinador."
            color={C.petrol}
            variant="decorated"
          />
          <FeatureCardCompact
            label="Panel de Control"
            description="Subida de archivos por arrastrar y soltar, configuración de filtros, seguimiento de workers activos con logs y un historial local persistente."
            color={C.petrol}
            variant="decorated"
          />
        </div>
        <div className="flex items-center justify-center gap-3">
          <span className="text-xl font-semibold uppercase tracking-[0.15em] text-gray-500">
            Stack
          </span>
          {["Astro", "React 19", "Tailwind v4", "shadcn/ui", "Bun", "Biome"].map(
            (s) => (
              <Badge key={s} label={s} color={C.petrol} />
            ),
          )}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── INFRAESTRUCTURA ─── */
function Infrastructure() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Infraestructura y Despliegue"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          De{" "}
          <span style={{ color: C.petrol }}>Docker Compose</span> a Kubernetes
        </h1>
        <div className="grid grid-cols-2 gap-3">
          <FeatureCardCompact
            label="Entorno local"
            description="Docker Compose orquesta Redis, MinIO, coordinador, reaper y workers, con capacidad de escalar réplicas horizontalmente."
            color={C.petrol}
            variant="decorated"
          />
          <FeatureCardCompact
            label="Entorno Kubernetes"
            description="Manifiestos de red, ingress y certificados, gestionados con Helm charts y Helmfile; secretos cifrados con SOPS."
            color={C.petrol}
            variant="decorated"
          />
        </div>
        <div className="grid grid-cols-4 gap-3">
          {[
            { v: "100%", l: "Contenerización multi-stage" },
            { v: "60%", l: "Observabilidad OpenTelemetry" },
            { v: "50%", l: "Despliegue Kubernetes / Helm" },
            { v: "40%", l: "Gestión de secretos con SOPS" },
          ].map((i) => (
            <StatCard
              key={i.l}
              value={i.v}
              label={i.l}
              color={C.petrol}
              variant="decorated"
            />
          ))}
        </div>
      </SlideWrap>
    </Slide>
  );
}

/* ─── EVIDENCIAS ─── */
function Evidence({
  index,
  title,
  highlight,
  desc,
  image,
}: {
  index: number;
  title: string;
  highlight: string;
  desc: string;
  image: string;
}) {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag={`Evidencias · ${index}/5`}
        variant="decorated"
        className="w-full flex flex-col gap-2 items-center justify-center"
      >
        <div className="flex flex-col items-center">
          <h1 className="text-4xl! text-balance text-center">
            {title}{" "}
            <span className="font-semibold" style={{ color: C.petrol }}>
              {highlight}
            </span>
          </h1>
          <p className="mt-1 text-xl text-gray-400 text-pretty text-center max-w-4xl">
            {desc}
          </p>
        </div>
        <img
          src={image}
          alt={`${title} ${highlight}`}
          className="object-contain rounded-xl bg-white/95 p-2 shadow-md max-h-[440px]"
        />
      </SlideWrap>
    </Slide>
  );
}

/* ─── TRABAJO PENDIENTE ─── */
function PendingWork() {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={C.petrol}
        tag="Trabajo Pendiente · Fase 5"
        variant="decorated"
        className="w-full flex flex-col gap-3 justify-center items-center"
      >
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Camino al{" "}
          <span style={{ color: C.petrol }}>cierre del proyecto</span>
        </h1>
        <div className="grid grid-cols-2 gap-3 self-stretch">
          {[
            {
              l: "Autenticación de usuarios",
              d: "RF-014 · Registro e inicio de sesión seguro para proteger datos y personalizar el acceso.",
            },
            {
              l: "Historial persistente",
              d: "RF-015 · Almacenamiento del historial de tareas en base de datos, accesible desde cualquier dispositivo.",
            },
            {
              l: "Pruebas automatizadas",
              d: "RNF-009 · Cobertura de pruebas unitarias e integración sobre los servicios del backend.",
            },
            {
              l: "Validación en clúster real",
              d: "Despliegue completo en Kubernetes y cierre de la observabilidad con Prometheus y Grafana.",
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
        <h1 className="text-5xl! font-semibold tracking-tight text-center">
          Un pipeline{" "}
          <span style={{ color: C.petrol }}>distribuido y funcional</span>
        </h1>
        <div className="grid grid-cols-2 gap-2">
          {[
            {
              n: "1",
              l: "Arquitectura validada",
              d: "El patrón productor-cola-consumidor con Redis y MinIO ya opera de forma estable en el entorno local.",
            },
            {
              n: "2",
              l: "Escalabilidad horizontal",
              d: "Los workers se replican sin cambios de código, tanto en Docker Compose como en Kubernetes.",
            },
            {
              n: "3",
              l: "Tolerancia a fallos",
              d: "El reaper y las colas atómicas (BLMOVE, SETNX) garantizan que ninguna tarea se pierda ante una caída.",
            },
            {
              n: "4",
              l: "Cierre en progreso",
              d: "Observabilidad, seguridad y pruebas automatizadas son las tareas centrales de la fase final.",
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
          Microphoto demuestra, con un avance del{" "}
          <span className="font-semibold text-white">50 %</span>, que el
          procesamiento distribuido de imágenes es{" "}
          <span className="font-semibold text-white">
            viable, resiliente y escalable
          </span>
          .
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
      <Introduction />
      <Team />
      <ProjectStatus />
      <Backlog />

      <Architecture />
      <SequencePart1 />
      <SequencePart2 />
      <Backend />
      <Frontend />
      <Infrastructure />

      <Evidence
        index={1}
        title="Página de inicio"
        highlight="informativa"
        desc="Landing page que explica el problema, el flujo del pipeline y las capacidades de la plataforma."
        image={IMG_EV_LANDING}
      />
      <Evidence
        index={2}
        title="Configuración"
        highlight="de filtros y parámetros"
        desc="Selección de filtro (grayscale, blur, brightness o resize) y sus parámetros antes de enviar la tarea."
        image={IMG_EV_CONFIG}
      />
      <Evidence
        index={3}
        title="Procesamiento"
        highlight="en tiempo real"
        desc="Progreso general de las fases del pipeline distribuido, actualizado mediante Server-Sent Events."
        image={IMG_EV_PROGRESS}
      />
      <Evidence
        index={4}
        title="Detalle"
        highlight="por nodo worker"
        desc="Vista expandida con el log individual de cada worker activo procesando sus fragmentos."
        image={IMG_EV_DETAIL}
      />
      <Evidence
        index={5}
        title="Resultado"
        highlight="final"
        desc="Previsualización de la imagen procesada y reconstruida, con opción de descarga."
        image={IMG_EV_RESULT}
      />

      <PendingWork />
      <Conclusions />
      <ThanksSlide color={C.petrol} variant="decorated" />
    </PresentationDeck>
  );
}
