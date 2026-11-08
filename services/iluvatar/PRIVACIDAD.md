# Inventario de datos personales de iluvatar

Documento técnico para quien redacte la política de privacidad y los Términos y Condiciones. **No es asesoría legal**: describe qué hace el código hoy.

## Qué datos trata el servicio

| Dato | Dónde queda | Para qué | Cuánto dura |
|---|---|---|---|
| `uid` de Firebase (opaco) | Solo en memoria durante la petición (`middleware.UID`) | Identificar la petición autenticada | La petición |
| Token de dispositivo (FCM) | No se guarda: se reenvía a Firebase para suscribirlo a topics | Notificaciones push | Lo que Firebase lo retenga |
| Topics suscritos (`all`, `service.<id>`) | En Firebase (asociados al token del dispositivo) | Elegir qué notificaciones recibir | Hasta que se dé de baja |
| Término de búsqueda | Solo en la caché en memoria del servicio (clave de la consulta), nunca en el log | Buscar publicaciones | Hasta 24 horas, con tope de 500 entradas |
| Dirección IP | Memoria del rate limit (3 minutos); no se escribe en el log | Limitar abusos | 3 minutos |

`iluvatar` no tiene base de datos propia. Los logs registran método, ruta registrada (no la URL con parámetros), estado, latencia e identificador de petición; nunca tokens, correos ni términos de búsqueda (hay un test que lo comprueba).

## Qué datos trata Firebase (Google)

Firebase Authentication guarda el correo (u otro identificador del proveedor de acceso), el proveedor y las fechas de acceso de cada cuenta; Firebase Cloud Messaging maneja los tokens de dispositivo. Ese tratamiento depende de la configuración del proyecto de Firebase y de sus condiciones.

## Qué falta decidir fuera del código

- Base legal y finalidad declaradas, y quién es el responsable del tratamiento.
- Cómo se ejercen los derechos de acceso, rectificación y eliminación (borrar la cuenta en Firebase Auth y dar de baja los topics).
- Si se vincula con Laurelin (registro de estudiantes): ahí sí habría datos personales propios y una evaluación aparte.
- El informe de la tesis calificaba los datos como "información privada de salud"; con el diseño actual (sin base de datos y sin preferencias en el servidor) esa clasificación debe revisarse.
