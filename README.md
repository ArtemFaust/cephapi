> Документация по работе с пайплайном
# Описание действий	связанных с работай пайплайна
## RGW операции

### Создание нового S3 пользователя
Для создания нового S3 пользователя необходимо
* Создать новый пайплайн
* Заполнить поля соответсвующим (указанным ниже) образом:
- **operation_type**: Выбор типа операции RGW RADOS RBD
	- Выбираем **RGW**
- **endpoint**: Ceph cluster endpoint
	- Выбираем конечную точку подключения. Класстер на котолром планируется создавать УЗ пользователя
- **operation**: Выбор операции
	- Выбираем соответсвующую нашей задачи операцию
		- **Создание нового S3 пользователя**
- **format**: Формат отображения. Применим только к операциям отображения данных
	- Для операции создания пользователя не имеет значения выбранный формат. Вы получаете тело письма которое нужно отправить пользователю.
- **user_display_name**: Отображаемое имя пользователя
	- Это не технический идентификатор пользователя. Имя пользователя отображаемое как его методанные. Создается из следующих параметров:
		- Контур - указан в заявке
		- Тенант - указан в заявке
		- RFS - номер заявки
	Пример:
		**prod.gismt.crpt.tech.RFS-114104**
- **user_email_address**: Email пользователя для которого создается учетная запись.
	- Так как пользователь может иметь более чем 1 УЗ а в рамках CEPH RGW для 1 email адресса может существовать толь ко одна УЗ - мы не указываем данный атрибут. Он сохранен для последующих доработок связанных с автоматической отправкой письма с данными доступа к учетной записи на электронную почту.
- **user_quota_gib**: Квота пользователя указывается в GiB - по умолчанию квота не устанавливается
	- Указывается запрошенная квота пользователем в GiB. 0 или -1 снимает ограничение на квоту. Если приходит заявка с 0 в поле квота значит ставим минимальную доступную квоту 1 GiB.
- **user_quota_obj**: Квота объектов пользователя указывается в кол-ве - по умолчанию квота не устанавливается
	- Оставляем по умолчанию. В рамках наших задач мы квоту не выставляем.
- **user_cups**: Permission пользователя.
	- В случае ЕСЛИ НЕОБХОДИМО переопределить дефолтьные свойства УЗ пользователя мы можем указать соответсвующие разрешения доступные пользователю. Формат заполнения соответсвует официальному формату работы с CEPH S3 с примерами котрого можно ознакомиться нижу в документации по работе с утилитой cephapi.
- **uuid**: UID пользователя S3
	- Технический идентификатор пользователя.
		- Если оставить пустым то UUID будет сгенерирован
		- Если нужно указать UUID вручную то указываем номер RFS. Пример: RFC-XXXXXXX

И того для заведения новой учетной записи в большинстве случаев достаточно указать:
- Кластер
- Отображаемое имя пользователя
- Квота в GiB
- UUID при необходимости

## RADOS операции
## Содание новой cephfs
Для создания новой cephfs необходимо
* Создать новый пайплайн
* Заполнить поля соответсвующим (указанным ниже) образом:
- **operation_type**: Выбор типа операции RGW RADOS RBD
	- Выбираем **RADOS**
- **endpoint**: Ceph cluster endpoint
	- Выбираем конечную точку подключения. Класстер на котолром планируется создавать новую cephfs
- **operation**: Выбор операции
	- Выбираем соответсвующую нашей задачи операцию
		- **Создание новой cephfs**
- **format**: Формат отображения. Применим только к операциям отображения данных
	- Для операции создания новой cephfs не имеет значения выбранный формат. Вы получаете тело письма которое нужно отправить пользователю.
- **cephfs_name**: Имя новой cephfs. 
	- За практику взято использовать в качетсве имени id rfs задачи
- **cephfs_placement**: Указание на каких нодах сервера будет распологаться mds демон
	- Нет крайней необходимости указывать конкретные сервера. Конкретные сервера имеет смысл указывать только в том случае если имется разделение области обслуживания сервисов на кластере (первые 3 обслуживают cephfs вторые 3 rgw и тд). У нас такой практики нет так что можно указывать все сервера через "*"
- **cephfs_meta_crush_rule**: Указание crush rule для размещения meta pool создаваемой cephfs. Если поле останется пустым конфигурация будет взята из дефолтных значений определенных на уровне конфигурации cephapi.
- **cephfs_data_crush_rule**: Указание crush rule для размещения data pool создаваемой cephfs. Если поле останется пустым конфигурация будет взята из дефолтных значений определенных на уровне конфигурации cephapi. Внимание запрос на предоставления доступа к cephfs может требовать указания конкретного типа носителей (ssd hdd и тд) именно тут в слущае необходимости мы должны указать требуемый crush rule для размещения data pool на требуемом типе носителей.
- **cephfs_data_pool_quota**: Установка ограничения на доступный объем к использованию создаваемой cephfs. Указывается в GiB.
- **cephfs_meta_pool_pg_num**: Установка количества PG для meta pool.
- **cephfs_data_pool_pg_num**: Установка количества PG для data pool.
- **cephfs_create_new_user**: Переключатель указывающий нужно ли создавать нового пользователя (owner) для создаваемой cephfs.
	- По умолчанию переклюатель установлен в **TRUE**.
	- По умолчанию имя создаваемого пользователя будет соответсвовать имени cephfs.
	- Если нужно задать конкретное имя создаваемого пользователя можно определить имя в соответсвующем input
- **cephfs_username**: Указание имени создаваемого cephfs пользователя. По умолчанию не заполнено и равно имени создаваемой cephfs.
	- Используется только если есть реальная необходимость определения имени пользователя отличного от названия cephfs.

## Смена crush rule существуещего pool cephfs
Для изменения crush rule существуещего pool cephfs необходимо
* Создать новый пайплайн
* Заполнить поля соответсвующим (указанным ниже) образом:
- **operation_type**: Выбор типа операции RGW RADOS RBD
	- Выбираем **RADOS**
- **endpoint**: Ceph cluster endpoint
	- Выбираем конечную точку подключения. Класстер на котолром планируется ыполнение операции.
- **operation**: Выбор операции
	- Выбираем соответсвующую нашей задачи операцию
		- **Смена crush_rule для существующего pool**
- **format**: Формат отображения. Применим только к операциям отображения данных
	- Для операции смены crush rule не имеет значения выбранный формат. Вы получаете лог выполнения.
- **cephfs_name**: Имя cephfs для pool которой будем проводить операцию смены crush rule.
- **cephfs_pool_type**: Тип пула для которого будем прододить смену crush rule.
	- Выбираем **data** 
		- **cephfs_data_crush_rule**: Указание crush rule для размещения data pool. Если поле останется пустым операция не будет выполненна.
	- Выбираем **meta**
		- **cephfs_meta_crush_rule**: Указание crush rule для размещения meta pool. Если поле останется пустым операция не будет выполненна.

> Документация к утилите

# Ceph api - cli интерфейс имплементации API

## <span style="color: yellow">Интерактивный режим</span>

Для входа в интерактивном режим используется флаг `-i`

## <span style="color: yellow">Сборка проекта</span>
- Создать `builder/.env` взяв за пример `builder/env.example`
- Указать ключ шифрования длиной 32 символа пример cfd8adfa-1df9-4b96-13er-6362e240
- Собрать проект для нужной архитектуры
```bash
cd builder
./runincontainer.sh build amd64 # x86_64  архитектура
./runincontainer.sh build arm64 # aarch64 архитектура
```

## <span style="color: yellow">Запуск из исходного кода без сборки</span>
```bash
# Указать соль
export CEPH_API_SALT="32 len key"
go run -ldflags "-X 'cephapi/utils.SALT=$CEPH_API_SALT'" . -i [-refresh] [-nocache]
```

## <span style="color: yellow">CEPH RGW - имплеминтация S3 протакола</span>

>Реализованный функционал
1. **Получение списка пользователей RGW или отдельного пользователя по его UID**

```bash
-rgw -lu -format <table|json> -e <endpoint name> -uid <user uid> [-refresh] [-nocache]
# -uid необязательный параметр. Выводит информацию по конкретному пользователю
# -size kb|gb|tb выводит данные размеров в указанном формате
```
```bash
UID                                   DISPLAY NAME                          USER TYPE  SUSPEND  MAX BUCKETS  QUOTA STATUS  QUOTA MAXSIZE  QUOTA MAX OBJECTS  BUCKETS COUNT  BUKETS ACTUAL SIZE  
---                                   ------------                          ---------  -------  -----------  ------------  -------------  -----------------  -------------  ------------------  
4228D396-A51A-43E3-BC3B-5AAF72004AB3  Test Testovich                        rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
D68417F9-62BF-4DC3-AA9C-F4507D1AF930  Test Testovich                        rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
3D326BEA-8FCE-4330-8C35-DA5D3996AD5E  Test Testovich                        rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
8FE7F860-5685-485C-BC8F-0ABD88655458  Test Testovich                        rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
436726c1-8b52-4ee7-a0a7-f4f9e71ee0b4  436726c1-8b52-4ee7-a0a7-f4f9e71ee0b4  rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
E0A04EDF-CA33-4B77-BF16-5B8603F0C401  Test Testovich                        rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         0              0 kb                
DE19485B-4B45-4CCB-B9C4-D4624AD90F3D  Test Testovich                        rgw        🟢 NO     1000         ✅ enabled     100 kb         100                0              0 kb     
```
2. **Получение списка бакетов** 

```bash
-rgw -lb -format <table|json> -e <endpoint name> -uid <user uid> [-refresh] [-nocache]
# -uid необязательный параметр. Выводит информацию по конкретному пользователю
# -size kb|gb|tb выводит данные размеров в указанном формате
```
```bash
OWNER  BUCKET         TENANT  QUOTA STATUS  QUOTA MAXSIZE  QUOTA MAX OBJECTS  BUKET ACTUAL SIZE  C TIME               M TIME               VERSIONED  VERSIONING  OBJECT LOCK  DATA POOL  INDEX POOL  SHARDS  
-----  ------         ------  ------------  -------------  -----------------  -----------------  ------               ------               ---------  ----------  -----------  ---------  ----------  ------  
artem  1234                   ❌ disabled    ❌ disabled     ❌ disabled         131988 kb          2026-02-10 17:05:00  2026-02-11 13:55:49  -          ✅ enabled   ✅ enabled                           11      
artem  4321                   ❌ disabled    ❌ disabled     ❌ disabled         0 kb               2026-02-10 17:05:04  2026-02-11 13:55:49  -          ✅ enabled   ✅ enabled                           11      
artem  5554                   ❌ disabled    ❌ disabled     ❌ disabled         0 kb               2026-02-10 17:05:08  2026-02-11 13:55:49  -          ✅ enabled   ✅ enabled                           11      
artem  asdasd                 ❌ disabled    ❌ disabled     ❌ disabled         0 kb               2026-02-11 13:47:35  2026-02-11 13:55:49  -          ✅ enabled   ✅ enabled                           11 
```

3. **Создание нового пользователя S3**

```bash
-rgw -cu \
	-rup "<USER ID>|<USER DISPLAY NAME>|<USER EEMAIL> or empty|<QUOTA SIZE IN Kb> or -1 or not set|<QUOTA OBJECTS> or -1 or not set|" \
	-cups "buckets=*;users=*;usage=read;metadata=read;zone=read" \
	-e <endpoint> -format table|json
# -cups - необязательный параметр если не задать пользователю не будут назначены CUPS
#   Указывается в формате buckets=*;users=*;usage=read;metadata=read;zone=read для нужных привелегий
# -rup - параметры создаваемого пользователя разделенные символом |
#   - UID пользователя|Отображаемое имя|Email|Размер квоты в Kb|Размер квоты на кол-во объектов|
```

3. **Создание SubUser**

```bash
-e ceph-single -rgw -format table -cus -uid a4f4073f-e80a-40bc-b416-1ac8fe5ae2b3 -rup "1|read"
UID                                     DISPLAY NAME                          USER TYPE  SUSPEND  MAX BUCKETS  QUOTA STATUS  QUOTA MAXSIZE  QUOTA MAX OBJECTS  BUCKETS COUNT  BUKETS ACTUAL SIZE  
---                                     ------------                          ---------  -------  -----------  ------------  -------------  -----------------  -------------  ------------------  
a4f4073f-e80a-40bc-b416-1ac8fe5ae2b3    a4f4073f-e80a-40bc-b416-1ac8fe5ae2b3  rgw        🟢 NO     1000         ❌ disabled    ❌ disabled     ❌ disabled         1              13028 kb            
a4f4073f-e80a-40bc-b416-1ac8fe5ae2b3:1  ➖ SU                                  ➖ SU       ➖ SU     ➖ SU         ➖ SU          ➖ SU           ➖ SU               ➖ SU           ➖ SU
# Где
# -rup "1|read" - 1 это порядковый номер subuser
# read - права пользователя. Доступны права: read, write, readwrite, full и none
```

4. Получение списка ключей и caps пользователя

```bash
-rgw -guk -uid <UID> -e <endpoint> -format table|json
```

5. **Установка или именение пользовательских CUPS**

```bash
-rgw -cuc -uid <UID> -caps "buckets=*;users=*" -e <endpoint>
```

6. **Удаление CAPS пользователя**

```bash
-rgw -ruc -uid <UID> -e <endpoint>
```

7. **Создания нового S3 ключа для пользователя**

Можно создать ключ для основную УЗ пользователя или через subuser.

```bash
-rgw -ck -e <enpoint> -uid <uid> -kt <s3|swift> -format <json|table> -suid <subuser uid>"
```

8. **Удаление subusers**

Удаление subuser пользователя s3

```bash
-rgw -rsu -e <endpoint> -uid <user uid> -suid <s_user uid> -format table
```

8. **Удаление S3 user**

Удаление пользователя s3

```bash
-rgw -ru -e <endpoint> -uid <user uid> -format table
```

9. **Смена Placement и Storage Class**
```bash
-rgw -e <endpoint> -uid <user uid> -cup -dp <placement_name> -ds <storage_class> -format table|json
```

## <span style="color: yellow">CEPH RADOS - имплеминтация rados протакола</span>
>Реализованный функционал
1. **Создание новой cephfs**
```bash
-rados -e <endpoint name> -cf -fsname <fs_name> -placement <server list or number> \
	-dmc <metadata pool crush_rule> \
	-ddc <data pool crush_rule> \
	-mpa "off|16|none|0|0" \
	-dpa "off|32|none|107374182400|0" \
# Обязательные параметры - fsname
# Описание параметров
# -cf - ключ указание операции создания новой cephfs
# -fsname - имя новой создаваемой cephfs
# -placement  - список серверов размещения. Пример - "osd0,osd1" или 3 или *. По умолчанию 3.
# -dmc - название crush_rule размещения metadata pool. Если не указанно берется из конфигурации DefaultMetaCrushRule для указанного endpoint
# -ddc - название crush_rule размещения data pool. Если не указанно берется из конфигурации DefaultDataCrushRule для указанного endpoint
# -mpa и -dpa- параметры создания metadata pool и data pool
# 	1. pg_autoscale on|off
# 	2. pg_num - количество PG
#	3. compresion - none|lz4|lz4hc
#	4. quota_bytes - Квота в байтах
#	5. quota_objects - Квота кол-ва объектов
```

2. **Смена crush_rule для существующего pool**
```bash
-rados -e <endpoint> -cpc -pool <pool name> -crush_rule <crush rule name>
# Пример
# -rados -e ceph-single -cpc -pool cephfs.testfs1.data -crush_rule replicated_rule
```

3. **Смена атрибутов существующего pool**
```bash
-rados -e <endpoint> -cpa -pool <pool name> -dpa "off|32|none|0|0" # Для data pool
-rados -e <endpoint> -cpa -pool <pool name> -mpa "off|32|none|0|0" # Для meta pool
# Пример
# -rados -e ceph-single -cpa -pool cephfs.testfs1.data -dpa "off|32|none|0|0"
```

4. **Создание нового CEPH USER**
```bash
-rados -e <endpoint> -cu -user <user entry name> -caps "mon==<value>;mds==<value>;osd==<value>"
# -user: может быть передан без приставки client. - в таком случая данная приставка будет добавленна к имени создаваемого пользователя
# -caps: это список разрешений для создаваемого пользователя. Не обязательный параметр - если не передавать пользователю не будут назначены ни какие права
# Пример:
#	-rados -e ceph-single -ccu -user client.RFS-0000 -caps "mon==allow *;mds==allow *; osd==allow *"
```

5. **Изменение или установка caps пользователя CEPH**
```bash
-rados -e <endpoint> -cuc -user <user entry name> -caps "mon==<value>;mds==<value>;osd==<value>"
# Пример:
# 	-rados -e ceph-single -cuc -user client.RFS-0000 -caps "mon==allow *;mds==allow *; osd==allow r"
```

6. **Получение информации о пользователе CEPH**
```bash
-rados -e <endpoint> -guk -user <user entry name>
# Пример:
# -rados -e ceph-single -guk -user client.RFS-0000
```

7. **Получить список существующих CephFS**
```bash
-rados -e <endpoint> -gf -format table|json -fsname <ceph fs name> [-refresh] [-nocache]
# -fsname - позволяет сократить вывод до укзанной cephfs
FSNAME          SUBVOLUMES    POOLS                       CACHE MODE PGNUM PG PENDING PG TARGET PGAUTOSCALE CRUSH RULE ID FLAGS           MAX BYTES   SIZE CREATE TIME      
------          ----------    -----                       ---------- ----- ---------- --------- ----------- ------------- -----           ---------   ---- -----------      
                              cephfs.ceph-single-nfs.meta none       32    32         32        on          0             hashpspool      0           2    2026-03-02 11:30 
ceph-single-nfs nfs02;nfs0... cephfs.ceph-single-nfs.data none       64    64         64        on          0             hashpspool,bulk 0           2    2026-03-02 11:30 
                              cephfs.testfs.meta          none       16    16         16        off         0             hashpspool      0           2    2026-03-11 11:39 
testfs                        cephfs.testfs.data          none       32    32         32        off         0             hashpspool,bulk 10737418240 2    2026-03-11 11:39 
                              cephfs.testfs1.meta         none       16    16         16        off         0             hashpspool      0           2    2026-03-11 13:36 
testfs1                       cephfs.testfs1.data         none       32    32         32        off         0             hashpspool,bulk 0           2    2026-03-11 13:36 
```

8. **Получить список всех subvolume указанной ephfFS**
```bash
-rados -e <endpoint> -gfs -format table|json -fsname <cephfs_name>
# Пример
#	-rados -e ceph-single -gfs -format json -fsname ceph-single-nfs
SUB NAME BYTES QUOTA BYTES USED USED PERCENT DATA POOL                   GID  ATIME               MTIME               
-------- ----------- ---------- ------------ ---------                   ---  -----               -----               
nfs02    1073741824  0          0.00         cephfs.ceph-single-nfs.data 1000 2026-03-13 12:13:22 2026-03-13 12:13:22 
nfs01    1073741824  0          0.00         cephfs.ceph-single-nfs.data 1000 2026-03-03 13:25:52 2026-03-04 06:45:19 
```

9. **Получить информацию о кластере**
```bash
-rados -e <endpoint> -ci -format table|json
ATTR               VALUE                                             
----               -----                                             
FSID               54960d7a-02b1-11f1-a1a1-000c29be26fe              
ADDRS              172.16.41.1:0/3930098780                          
C NET                                                                
P NET                                                                
INSTANCE ID        1032473                                           
MON/S              [v2:172.16.41.129:3300/0,v1:172.16.41.129:6789/0] 
SIZE               20963328                                          
AVALIABLE          18951376                                          
USED               2011952                                           
OBJECTS            621                                               
MON MAX PG PER OSD 1000                                              
MON ALLOW POOL DEL false                                             
PG AUTOSCALE       on                                                
STS                true                                              
STS KEY            770a74e4006cb77d3b2dcda4b919e4ff
```

10. **Удаление пользователя ceph**
```bash
-rados -e <endpoint> -ru -user <username>
```
