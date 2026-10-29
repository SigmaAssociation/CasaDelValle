# Casa del Valle

## Documento de cierre del proyecto

Universidad de San Carlos de Guatemala · CUNOC · Teoría de Sistemas 1 · Grupo 4

29 de octubre de 2026

Integrantes: Josecarlos Emanuel Hernández Chuvac, Mauricio Joel Gómez Barrios, Mynor Daniel Morales Arriaza, Francisco Ronaldo Hilario García y Christian Anthony Gutiérrez Paxtor.

Contenido: requerimientos, diagramas de arquitectura y procesos; historial por sprint y retrospectiva; estimaciones y evidencia; fuentes. Guion y presentación se entregan por separado.

# Documento de requerimientos

## Propósito

Casa del Valle centraliza la publicación y reserva de cabañas, la disponibilidad por fechas y el registro del ciclo de estadía y sus movimientos financieros. Busca que huésped y anfitrión consulten la misma información de reservas, anticipos, pagos y avisos.

## Alcance y actores

El sistema es una aplicación web Angular con API Go y PostgreSQL, desplegada mediante Docker en una VM de Azure. Incluye cuentas, perfiles, cabañas e imágenes, reservas, anticipos, pagos simulados, saldos, check-in/out, cancelaciones, reembolsos, avisos y reportes. Temporadas y mantenimiento forman parte del objetivo de cierre, con diferencias de integración detalladas abajo.

| Actor | Responsabilidad |
| --- | --- |
| Huésped | Gestionar perfil, consultar cabañas, reservar, anticipar, abonar, cancelar, consultar movimientos, avisos y decidir ante mantenimiento. |
| Anfitrión | Gestionar propiedades propias, imágenes, tarifas, reservas y llegada/salida; consultar reportes autorizados y mantenimientos. También puede reservar como huésped. |
| Administrador | Administrar funciones autorizadas de usuarios y temporadas globales; consultar reportes. |

La base almacena un solo `id_rol` por usuario. Un administrador propietario puede usar mantenimiento por su identidad de dueño; esto no implica que el esquema soporte múltiples roles simultáneos.

Quedan fuera del alcance acreditado los cobros bancarios reales, pasarelas, devoluciones bancarias efectivas, correo/WhatsApp, alta disponibilidad y una aplicación móvil nativa. Las tablas de comentarios, valoraciones y comisiones existen, pero no se acredita un flujo completo de producto para ellas. La rúbrica inicial menciona hotel/habitaciones/recepcionista; el producto entregado usa cabañas y los tres roles anteriores. Falta adjuntar la validación histórica de esa adaptación de alcance.

## Identificación y trazabilidad

Se usan IDs `RF-001` a `RF-020` porque varios títulos del backlog repiten IDs `US-002`, `US-004` o `US-005`. El identificador estable del backlog es el número de issue. **Presente** significa observado en código, no aceptado funcionalmente en esta revisión.

<a id="rf-001"></a>

### **RF-001** Registro y autenticación

**Backlog:** [#14](https://github.com/SigmaAssociation/CasaDelValle/issues/14), [#32](https://github.com/SigmaAssociation/CasaDelValle/issues/32), [#33](https://github.com/SigmaAssociation/CasaDelValle/issues/33), [#34](https://github.com/SigmaAssociation/CasaDelValle/issues/34), [#36](https://github.com/SigmaAssociation/CasaDelValle/issues/36), [#37](https://github.com/SigmaAssociation/CasaDelValle/issues/37), [#63](https://github.com/SigmaAssociation/CasaDelValle/issues/63).

**Criterio de aceptación:** Crear cuenta con datos válidos, impedir correo/DPI duplicados, almacenar contraseña con bcrypt y permitir inicio con JWT. Rechazar credenciales inválidas y tokens vencidos.

<a id="rf-002"></a>

### **RF-002** Perfil y autorización

**Backlog:** [#30](https://github.com/SigmaAssociation/CasaDelValle/issues/30), [#45](https://github.com/SigmaAssociation/CasaDelValle/issues/45), [#46](https://github.com/SigmaAssociation/CasaDelValle/issues/46), [#114](https://github.com/SigmaAssociation/CasaDelValle/issues/114).

**Criterio de aceptación:** Editar datos del perfil respetando permisos. Aplicar roles Administrador, Huésped y Anfitrión en servidor y vistas. El anfitrión conserva funciones de huésped.


<a id="rf-003"></a>

### **RF-003** Catálogo de cabañas

**Backlog:** [#50](https://github.com/SigmaAssociation/CasaDelValle/issues/50), [#51](https://github.com/SigmaAssociation/CasaDelValle/issues/51), [#52](https://github.com/SigmaAssociation/CasaDelValle/issues/52), [#65](https://github.com/SigmaAssociation/CasaDelValle/issues/65), [#64](https://github.com/SigmaAssociation/CasaDelValle/issues/64), [#66](https://github.com/SigmaAssociation/CasaDelValle/issues/66), [#73](https://github.com/SigmaAssociation/CasaDelValle/issues/73), [#75](https://github.com/SigmaAssociation/CasaDelValle/issues/75).

**Criterio de aceptación:** Registrar, consultar, buscar, editar y eliminar cabañas según permisos; validar nombre, dirección, capacidad y precio. Mostrar mensajes ante datos inválidos o acceso ajeno.


<a id="rf-004"></a>

### **RF-004** Imágenes de cabañas

**Backlog:** [#92](https://github.com/SigmaAssociation/CasaDelValle/issues/92), [#96](https://github.com/SigmaAssociation/CasaDelValle/issues/96), [#97](https://github.com/SigmaAssociation/CasaDelValle/issues/97), [#98](https://github.com/SigmaAssociation/CasaDelValle/issues/98), [#102](https://github.com/SigmaAssociation/CasaDelValle/issues/102).

**Criterio de aceptación:** Agregar, consultar y eliminar imágenes asociadas a una cabaña con autorización sobre las operaciones de escritura. Mostrar fotografías en catálogo/detalle.


<a id="rf-005"></a>

### **RF-005** Reserva y disponibilidad

**Backlog:** [#99](https://github.com/SigmaAssociation/CasaDelValle/issues/99), [#100](https://github.com/SigmaAssociation/CasaDelValle/issues/100), [#109](https://github.com/SigmaAssociation/CasaDelValle/issues/109), [#187](https://github.com/SigmaAssociation/CasaDelValle/issues/187).

**Criterio de aceptación:** Aceptar inicio anterior a fin y no anterior a hoy. Impedir conflictos con reservas que bloquean disponibilidad y con otras reservas del huésped. Permitir salida e ingreso contiguos. Consultar calendario disponible.


<a id="rf-006"></a>

### **RF-006** Anticipo y confirmación

**Backlog:** [#158](https://github.com/SigmaAssociation/CasaDelValle/issues/158), [#159](https://github.com/SigmaAssociation/CasaDelValle/issues/159), [#186](https://github.com/SigmaAssociation/CasaDelValle/issues/186).

**Criterio de aceptación:** Mantener reserva pendiente hasta registrar anticipo de al menos 30 % del alojamiento y no superior al total. Guardar movimiento y confirmar la disponibilidad otra vez antes de confirmar.

<a id="rf-007"></a>

### **RF-007** Pagos y consulta de saldo

**Backlog:** [#160](https://github.com/SigmaAssociation/CasaDelValle/issues/160), [#161](https://github.com/SigmaAssociation/CasaDelValle/issues/161), [#165](https://github.com/SigmaAssociation/CasaDelValle/issues/165).

**Criterio de aceptación:** Registrar abonos simulados mayores que cero y no superiores al saldo, con método, fecha y reserva. Permitir consulta de pagos y estado. Rechazar pagos en estados no admitidos.


<a id="rf-008"></a>

### **RF-008** Check-in y check-out

**Backlog:** [#163](https://github.com/SigmaAssociation/CasaDelValle/issues/163), [#164](https://github.com/SigmaAssociation/CasaDelValle/issues/164).

**Criterio de aceptación:** Registrar llegada dentro del periodo permitido y con alojamiento pagado. Permitir cubrir el saldo exacto en efectivo al llegar. Registrar salida solo si la reserva está en curso y tiene check-in.

<a id="rf-009"></a>

### **RF-009** Cancelación y reembolsos ordinarios

**Backlog:** [#101](https://github.com/SigmaAssociation/CasaDelValle/issues/101), [#110](https://github.com/SigmaAssociation/CasaDelValle/issues/110), [#167](https://github.com/SigmaAssociation/CasaDelValle/issues/167).

**Criterio de aceptación:** Registrar cancelación y devolución como movimiento nuevo. Política del código: 100 % con 7 o más días, 70 % entre 3 y 6 días y 0 % con menos de 3 días. No devolver más del dinero disponible. Pagos y reembolsos son simulados.


<a id="rf-010"></a>

### **RF-010** Buzón personal

**Backlog:** [#113](https://github.com/SigmaAssociation/CasaDelValle/issues/113).

**Criterio de aceptación:** Consultar avisos propios en orden descendente, contador de no leídos y marcado de lectura. Avisar eventos de reservas y pagos.


<a id="rf-011"></a>

### **RF-011** Reporte de reservaciones

**Backlog:** [#166](https://github.com/SigmaAssociation/CasaDelValle/issues/166).

**Criterio de aceptación:** Consultar información real de reservas con filtros autorizados y estado vacío cuando no existan resultados. Mostrar cantidad de cabañas y reservaciones.


<a id="rf-012"></a>

### **RF-012** Temporadas globales

**Backlog:** [#194](https://github.com/SigmaAssociation/CasaDelValle/issues/194), [#196](https://github.com/SigmaAssociation/CasaDelValle/issues/196).

**Criterio de aceptación:** Crear, listar, actualizar y desactivar temporadas globales con rango inclusivo y porcentaje sugerido. Impedir cruces entre temporadas globales activas. Reservar escritura a administrador.


<a id="rf-013"></a>

### **RF-013** Tarifas personalizadas

**Backlog:** [#193](https://github.com/SigmaAssociation/CasaDelValle/issues/193), [#195](https://github.com/SigmaAssociation/CasaDelValle/issues/195).

**Criterio de aceptación:** Gestionar precio fijo o porcentaje por cabaña y periodo. Mostrar precio recomendado a partir de la base y porcentaje global. Poder eliminar tarifa y reutilizar su periodo. Gestionar precios desde apartado propio.


<a id="rf-014"></a>

### **RF-014** Cotización y total persistido

**Backlog:** [#193](https://github.com/SigmaAssociation/CasaDelValle/issues/193), [#198](https://github.com/SigmaAssociation/CasaDelValle/issues/198).

**Criterio de aceptación:** Sumar precio de cada noche de [inicio, fin), aplicar regla particular antes que global y luego base. Guardar precio_total al crear reserva y reutilizarlo en anticipos, pagos y saldo.


<a id="rf-015"></a>

### **RF-015** Bloqueo por mantenimiento

**Backlog:** [#197](https://github.com/SigmaAssociation/CasaDelValle/issues/197), [#198](https://github.com/SigmaAssociation/CasaDelValle/issues/198).

**Criterio de aceptación:** El propietario, incluso si su rol es administrador, puede previsualizar reservas afectadas y bloquear su cabaña por fechas y motivo. No permitir nuevas reservas en el intervalo. Suspender reservas afectadas y avisar al huésped.


<a id="rf-016"></a>

### **RF-016** Decisión del huésped afectado

**Backlog:** [#198](https://github.com/SigmaAssociation/CasaDelValle/issues/198).

**Criterio de aceptación:** El huésped elige nuevas fechas disponibles en la misma cabaña, conservando noches, total original y pagos, o solicita devolver todo lo pagado pendiente de devolución sin penalización. La resolución no puede aplicarse dos veces.


<a id="rf-017"></a>

### **RF-017** Cancelación del mantenimiento

**Backlog:** [#198](https://github.com/SigmaAssociation/CasaDelValle/issues/198).

**Criterio de aceptación:** El propietario puede cancelar el bloqueo en cualquier momento. Restaurar reservas aún utilizables y sin decisión si ya no hay otro bloqueo. Avisar al huésped. Conservar historial y decisiones ya resueltas.


<a id="rf-018"></a>

### **RF-018** Reporte de ingresos

**Backlog:** [#203](https://github.com/SigmaAssociation/CasaDelValle/issues/203).

**Criterio de aceptación:** Filtrar movimientos reales por fechas y opcionalmente cabaña. Restringir anfitrión a propiedades propias. Mostrar pagos, reembolsos e ingreso neto, validar rango y mostrar total cero sin resultados.


<a id="rf-019"></a>

### **RF-019** Porcentaje de cancelación

**Backlog:** [#208](https://github.com/SigmaAssociation/CasaDelValle/issues/208).

**Criterio de aceptación:** Mostrar porcentaje de reservas canceladas respecto al total seleccionado y cero cuando el total sea cero. Respetar filtros y permisos del reporte.


<a id="rf-020"></a>

### **RF-020** Migraciones y operación

**Backlog:** [#188](https://github.com/SigmaAssociation/CasaDelValle/issues/188).

**Criterio de aceptación:** Aplicar SQL pendiente con script desde Docker de desarrollo, conservar historial por archivo, omitir aplicadas, coordinar ejecuciones simultáneas y detener ante error. Documentar el procedimiento.


## Reglas de negocio que deben mantenerse coherentes

1. Reservas y mantenimientos usan **inicio, fin**: no se cobra la noche de salida y dos intervalos contiguos no se cruzan. Las temporadas usan extremos **inclusivos**.
2. En main, disponibilidad comprueba principalmente reservas `confirmada` y `en_curso`. Una reserva `pendiente` no equivale a disponibilidad garantizada. Al pagar anticipo se comprueba otra vez. La prueba concurrente sigue pendiente.
3. El anticipo mínimo del código es **30 %**, fijo. La ficha inicial hablaba de porcentaje configurable; no debe describirse como configuración disponible.
4. El check-in manual exige alojamiento liquidado. El check-out exige estado `en_curso` y fecha de check-in. Existe también avance automático por fechas que puede cambiar estados sin verificar el pago completo y completar marcas históricas; no debe confundirse con la llegada física validada.
5. `precio_total` es el total del **alojamiento**. El saldo incorpora por separado cargos, pagos y reembolsos. Una devolución total por mantenimiento devuelve dinero cobrado aún no devuelto, no el valor nominal de noches impagadas.
6. Prioridad de precio: tarifa particular fija, tarifa particular porcentual (también 0 %), temporada global activa, tarifa base. Se redondea cada noche a dos decimales y se suman las noches. Ejemplo de aceptación, no dato real: base Q100, una noche ordinaria y dos al +25 % dan Q350 y anticipo mínimo Q105.
7. Reprogramar por mantenimiento conserva cabaña y cantidad de noches. Conserva el total aunque las nuevas fechas tengan otra temporada. Esto amplía la redacción original de #198 y responde al cambio pedido.

## Requerimientos no funcionales medibles


| ID | Medida y condición de aceptación | Método y estado | Trazabilidad |
| --- | --- | --- | --- |
| RNF-001 Seguridad de acceso | El 100 % de casos sin JWT, con JWT vencido y de acceso a recursos ajenos de la matriz de permisos debe recibir 401/403, sin modificar datos. | Pruebas API por rol y propietario. Middleware y controles presentes; ejecución de matriz pendiente. | RF-001/002/015; #63, #114. |
| RNF-002 Contraseñas y sesión | Ninguna contraseña nueva se guarda en claro. Todas se verifican con bcrypt. JWT emitido con expiración de 24 h según código. | Inspección de servicio y prueba de registro/login con usuario de demo. Inspección realizada; prueba pendiente. | RF-001; #14, #32. |
| RNF-003 Integridad | Cero reservas activas aceptadas durante mantenimiento y como máximo una confirmación exitosa en un ensayo de 20 solicitudes concurrentes para la misma cabaña/fechas. | Prueba sobre PostgreSQL aislado con el esquema final. Objetivo sin ejecutar; no se acredita exclusión concurrente general solo por comprobar disponibilidad en servicio. | RF-005/015/016. |
| RNF-004 Exactitud monetaria | Diferencia de Q0.00 entre total esperado, cotización, precio persistido, anticipo y saldo en los casos base, cruce de temporada, tarifa particular 0 % y reprogramación. | Casos con importes conocidos y revisión de movimientos. Pendiente en main final. | RF-006/007/014/016. |
| RNF-005 Rendimiento | p95 de catálogo y disponibilidad ≤2 s con 20 usuarios concurrentes, 1 000 reservas, durante 5 min; errores HTTP 5xx <1 %. Registrar capacidad de la VM. | Medición contra entorno acordado con datos sintéticos. Sin medición disponible. | RF-003/005; objetivo propuesto. |
| RNF-006 Despliegue verificable | Cada PR a main/develop ejecuta pruebas Go, Angular y comprobación de Swagger; cada push a main publica imágenes y despliega el mismo SHA. | Workflows inspeccionados. CD 40910471125 exitoso para la base main; falta validar el SHA final y la aceptación posterior. | #112; C03/C04 en evidencias. |
| RNF-007 Recuperación | Restaurar copia en base aislada en ≤30 min, con coincidencia de conteos y cinco reservas/movimientos de control. | Ensayo cronometrado de backup/restore. Pendiente de evidencia. No restaurar sobre producción durante la demo. | Rúbrica Sprint 6; RF-020. |
| RNF-008 Interfaz | Completar registro, reserva, anticipo y consulta de saldo sin desbordamiento horizontal a 360×800 y 1366×768; errores junto al formulario y acciones legibles. | Recorrido manual en ambas resoluciones y capturas del SHA final. Pendiente. | RF-001/005/006/007/013. |
| RNF-009 Persistencia y migraciones | Reiniciar servicios conserva reservas, movimientos e imágenes; ejecutar dos veces migraciones no duplica entradas ni cambios. Ante fallo no se registra el archivo fallido. | Ensayo en volumen de desarrollo desechable. Script inspeccionado; evidencia de ensayo de cierre pendiente. | RF-020; #188. |

## Restricciones y supuestos

**Restricciones confirmadas:** Angular, Go y PostgreSQL en contenedores; exposición de máximo diez minutos; documentos en Markdown, entrega en `main`; pagos simulados; Azure autorizado por el docente. Los porcentajes y plazos ordinarios son los del código, no de una política bancaria externa.


## Cambios del Product Owner y decisiones de cierre

| Decisión | Fuente y efecto verificable | Situación |
| --- | --- | --- |
| Reserva requiere compromiso económico | Rúbrica Sprint 5 y #159: anticipo mínimo 30 %, reserva pendiente hasta pago. | Código presente en main. |
| Mantenimiento afecta reservas y puede cancelarse | #198 y rúbrica Sprint 6. | PR #214 en develop al consultar. |
| El huésped decide la solución | Responsable del proyecto en esta conversación: elegir nuevas fechas en la misma cabaña o devolución total. Conservación del total original. | Actualizar criterio del backlog para reflejar esta versión. |
| Total persistido con temporadas | Responsable primero aplazó el total y después pidió añadirlo. La decisión vigente es guardar `precio_total` calculado por noche. | no describir la decisión anterior como vigente. |
| Tarifas fuera de editar cabaña | Solicitud del responsable para apartado independiente con precio recomendado y mejor legibilidad. |  |
| Azure para entrega | Autorización docente confirmada por el responsable en esta conversación. | Sustituye EC2 como restricción de infraestructura para esta entrega. |

## Equipo y reparto de responsabilidades

La ficha de arranque es la fuente de nombres y especialidades. El equipo confirmó que los roles se conservaron y que todos participaron en desarrollo. Las asignaciones del backlog respaldan áreas de trabajo, no porcentajes de esfuerzo.

| Integrante | Especialidad inicial conservada | Ejemplos de evidencia de desarrollo |
| --- | --- | --- |
| Josecarlos Emanuel Hernández Chuvac | Scrum Master inicial y de Sprint 1 | Cuenta Emahch: reservas BE #100, JWT #63, check-in #163 y mantenimiento #198. |
| Mauricio Joel Gómez Barrios | Base de datos 1 | Maurelo01: registro BE #33, perfil/cabañas #64, anticipos #159, reembolsos #167 e ingresos #203. |
| Mynor Daniel Morales Arriaza | Infraestructura | MynorDaniel: Docker #5, dominio #38, pruebas/despliegue #112, saldo #160, check-out #164 y precios #193. |
| Francisco Ronaldo Hilario García | Base de datos 2 | ronaldo901000: ER #3/#189, registro FE #34, perfil BE #45, cancelación #101 y reporte #166. |
| Christian Anthony Gutiérrez Paxtor | Repositorio | AnthonyGup: repositorio #12/#27, login FE #37, cabañas BE #51, pagos #161 y disponibilidad #187. |

No se acredita un Product Owner integrante, una rotación posterior de Scrum Master ni autoría exclusiva por módulo. No deben añadirse sin confirmación.

## Retrospectiva general

**Qué funcionó, confirmado por el equipo.** Asignar a una misma persona el backend y frontend de una funcionalidad redujo dudas al consumir sus endpoints: quien implementaba la API sabía cómo utilizarla en la interfaz. Dividir por funcionalidades completas dio una unidad de responsabilidad que el equipo podía revisar de extremo a extremo.

**Qué no funcionó, confirmado por el equipo.** Repartir el trabajo únicamente mediante roles específicos no cubría la necesidad de que todos desarrollaran. Los roles se mantuvieron como especialidades de apoyo y cada integrante participó también en implementación.

**Qué muestra el repositorio.** Hubo correcciones posteriores de integración, autorización, SSH, pagos y migraciones. Esto respalda que la integración y la operación exigieron trabajo adicional; no permite afirmar cuántas horas se perdieron. La existencia de CI/CD, PR y pruebas dio puntos verificables de revisión, aunque falta confirmar controles de rama y aceptación completa de cada incremento.

**Lecciones y cambios propuestos.** Mantener responsables por funcionalidad completa y pedir revisión cruzada para evitar conocimiento concentrado. Acordar contratos de API y ejemplos antes de implementar interfaz. Incluir migración, pruebas, documentación y aceptación por rol dentro de la tarea, en vez de tratarlos como cierre posterior. Al cambiar una regla de precio, validar creación, anticipo, abonos, saldo, reembolsos y reportes en el mismo recorrido.
