# Stock Tracker API

Una API para realizar un seguimiento de acciones en tiempo real, desarrollada con Go y Gin, optimizada para ejecutarse en una Raspberry Pi Zero 2W.

## Características

- Registro de compra de acciones con precio en tiempo real
- Seguimiento de ganancias/pérdidas
- Interfaz Swagger para documentación y pruebas
- Optimizado para bajo consumo de recursos

## Requisitos

- Go 1.21 o superior
- MongoDB Atlas (o servidor MongoDB local)
- Cuenta en [Alpha Vantage](https://www.alphavantage.co/) para obtener la API key

## Instalación

1. Clonar el repositorio:
   ```bash
   git clone [URL_DEL_REPOSITORIO]
   cd stock-tracker
   ```

2. Configurar variables de entorno:
   - Copiar el archivo `.env.example` a `.env`
   - Configurar las variables en `.env` con tus credenciales

3. Descargar dependencias:
   ```bash
   go mod tidy
   ```

## Configuración

Crear un archivo `.env` en la raíz del proyecto con las siguientes variables:

```
MONGO_URI=tu_cadena_de_conexion_mongodb
API_KEY_ALPHAVANTAGE=tu_api_key_de_alphavantage
```

## Uso

1. Iniciar la aplicación:
   ```bash
   go run main.go
   ```
   O compilar y ejecutar:
   ```bash
   go build -o stock-tracker
   ./stock-tracker
   ```

## Endpoints

- `GET /api/stocks`: Obtiene todas las acciones en la cartera
- `POST /api/stocks`: Agrega una nueva acción a la cartera
- `GET /api/stocks/summary`: Obtiene un resumen de la cartera con ganancias/pérdidas

## Ejemplo de uso

### Agregar una acción
```bash
curl -X POST "http://localhost:8080/api/stocks" \
     -H "Content-Type: application/json" \
     -d '{"symbol":"AAPL", "quantity": 5}'
```

### Ver todas las acciones
```bash
curl "http://localhost:8080/api/stocks"
```

### Ver resumen de la cartera
```bash
curl "http://localhost:8080/api/stocks/summary"
```

## Despliegue con Docker en Raspberry Pi Zero 2W

### Requisitos previos

1. Instalar Docker y Docker Compose en la Raspberry Pi:
   ```bash
   # Instalar Docker
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   
   # Agregar el usuario actual al grupo docker
   sudo usermod -aG docker $USER
   
   # Instalar Docker Compose
   sudo apt-get install -y libffi-dev libssl-dev
   sudo apt-get install -y python3 python3-pip
   sudo apt-get remove python-configparser
   sudo pip3 install docker-compose
   
   # Reiniciar la sesión para aplicar los cambios
   newgrp docker
   ```

2. Clonar el repositorio:
   ```bash
   git clone [URL_DEL_REPOSITORIO]
   cd stock-tracker
   ```

### Configuración

1. Editar el archivo `.env` si es necesario (por defecto ya viene con una configuración funcional para desarrollo local)

### Iniciar la aplicación

```bash
# Construir y levantar los contenedores
docker-compose up -d --build

# Ver los logs
docker-compose logs -f
```

### Comandos útiles

- Detener los contenedores:
  ```bash
  docker-compose down
  ```

- Ver los logs de la aplicación:
  ```bash
  docker-compose logs -f app
  ```

- Ver los logs de MongoDB:
  ```bash
  docker-compose logs -f mongo
  ```

- Acceder a la consola de MongoDB:
  ```bash
  docker-compose exec mongo mongosh -u root -p example
  ```

- Hacer backup de la base de datos:
  ```bash
  docker-compose exec -T mongo mongodump --archive --gzip --db=stock_tracker --username=root --password=example > backup_$(date +%Y%m%d_%H%M%S).gz
  ```

### Actualizar la aplicación

```bash
# Detener los contenedores
docker-compose down

# Obtener los últimos cambios
git pull

# Reconstruir y levantar los contenedores
docker-compose up -d --build
```

### Configuración para producción

1. Cambiar las credenciales de MongoDB en el archivo `.env`
2. Configurar un proxy inverso como Nginx
3. Configurar SSL con Let's Encrypt
4. Configurar copias de seguridad automáticas

## Licencia

MIT
