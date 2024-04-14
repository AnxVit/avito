# Тестовое задание. Avito
## Команды сборки
Посмотреть все команды:
```
make help
```

1. Для запуска приложения нужно прописать команду `make run`.

2. Для запуска приложения в контейнере: `make d.up`

3. Для запуска тестов: `make test`

## API

### GET /user_banner?tag_id={}&feature_id={}&use_last_version={}

    - Header: token

    - Return: banner:JSON

    Handler: `userbanner.NewGet(...)`

    DB:      `GetUserBanner(tag, feature, admin) (banner, error)`

### GET /banner?tag_id={}&feature_id={}&limit={}&offset={}

    - Header: token

    - Return: banners:[]JSON

    Handler: `banner.NewGet(...)`

    DB:      `GetBanner(tag, feature, limit, offset) ([]banner, error)`


### POST /banner

    - Header: token

    - Body:
    {

        "tag_ids": [int],

        "feature_id": int,

        "content": JSON,

        "is_active": bool

    }
    
    - Return: id:int

    Handler: `banner.NewPost(...)`

    DB:      `PostBanner(banner) (id, error)`

### PATCH /banner/{id}

    - Header: token

    - Body:
    {

        "tag_ids": [int]    `nullable`,

        "feature_id": int   `nullable`,

        "content": JSON     `nullable`,

        "is_active": bool   `nullable

    }
    
    Handler: `banner.NewPatch(...)`

    DB:      `PatchBanner(id, bannerPost) (error)`

### DELETE /banner/{id}

    - Header: token

    Handler: `banner.NewDelete(...)`

    DB:      `DeleteBanner(id) (error)`


## Примеры использования
### Создание банера
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/post.jpg?raw=true)
### Получение банера от имени пользователя
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/user_get.jpg?raw=true)
### Изменяем данные
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/patch.jpg?raw=true)
### Получаем данные от имени пользователя
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/user_get2.jpg?raw=true)
### Получаем данные от имени пользователя с флагом use_last_revesion
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/user_get_last.jpg?raw=true)

### Удаляем баннер
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/delete.jpg?raw=true)
### Пытаемся получить баннер после удаления
![Alt text](https://github.com/AnxVit/avito/tree/main/photo/get_delete.jpg?raw=true)

## Дополнительная реализация
- Была настроена конфигурация линтера `golangci-lint` в файле `.golangci.yml`

    Также был добавлен Left Hook для контроля пре-коммита: вызов линтера

- Интеграционные тесты были написаны для всех методов
    
    Были использованы такие библиотеки как:
    - `testcontainer-go`      : для создания базы данных в среде контейнера
    - `testfixtures`          : для создания тестовых данных; тестовые данные находятся в папке test/fixtures/storage
    - `stretchr/testify/suite`: для более удобного создания тестов
- Нагрузочное тестирование
    Использовался инструмент Bombardir
    На тестирвование были взяты: Get /user_banner и Get /banner
    Тесты запросы Get /banner с количеством запросов 1'000'000 и макисальным кол-вом одновмеременных подключений 100

    ![Alt text][def1]

    Тесты запросы Get /user_banner с количеством запросов 1'000'000 и макисальным кол-вом одновмеременных подключений 500

    ![Alt text][def2]

## Вопросы и проблемы
Проблемы возникли при создании кэша, который мог бы выдавать пользователям устаревшие баннеры.
Было много решений одно из низ: использовать кастомное хранилище с Redis. Однако я остановился
на варианте простого кастомного хранилища с использованием sync.Map и debouncer, который отвечает
за жизнь и обновление жизни.

В качестве токенов были использованы три вида: user_token, admin_token и noaccess.
Они выводились в Middleware Auth.

Были проблемы с логикой создания и обновления банеров.

- Patch: пользователь мог не указывать, а мог указать значения null. Это совершенно два разных варианта.

    В первом случае, те поля, которые были пропущены, не изменяться. А во втором, будут принимать значение null.
Добиться такого эффекта, можно было с помощью обертки под тип Optional.

- Post: админ должен отправить все параметры, так как, по логике, публикация без каких-то полей не имеет смысла.

    Пока эти поля не будут известны, не имеет смысла "постить" баннер

Админ полностью несет ответственность за создание banner

На первом этапе была построена одна таблица banner, а позже расширена на 4 таблицы: tag, feature, banner, bannertag.

Где bannertag была связью между banner и tag (связь многие-ко-многим). Banner и feature - связь один ко многим.  

[def1]: https://github.com/AnxVit/avito/tree/main/photo/bench_admin.jpg?raw=true
[def2]: https://github.com/AnxVit/avito/tree/main/photo/bench_user.jpg?raw=true