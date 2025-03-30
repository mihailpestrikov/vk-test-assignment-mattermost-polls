box.cfg{
    listen = '3301',
}


print("Starting initialization script init.lua")

local function bootstrap()
    if box.space.polls then box.space.polls:drop() end
    if box.space.votes then box.space.votes:drop() end

    local polls = box.schema.space.create('polls', {
        if_not_exists = false,
        format = {
            {name = 'id', type = 'string'},            -- ID голосования
            {name = 'question', type = 'string'},      -- Вопрос
            {name = 'options', type = 'array'},        -- Варианты ответов
            {name = 'created_by', type = 'string'},    -- ID создателя
            {name = 'channel_id', type = 'string'},    -- ID канала
            {name = 'created_at', type = 'number'},    -- Unix timestamp создания
            {name = 'expires_at', type = 'number'},    -- Unix timestamp истечения срока
            {name = 'status', type = 'string'}         -- Статус (ACTIVE, CLOSED, DELETED)
        }
    })

    -- По ID голосования (первичный)
    polls:create_index('primary', {
        type = 'HASH',
        unique = true,
        parts = {'id'},
        if_not_exists = true
    })

    -- По статусу и времени завершения (для автозавершения)
    polls:create_index('status_expires', {
        type = 'TREE',
        unique = false,
        parts = {'status', 'expires_at'},
        if_not_exists = true
    })

    -- По каналу (для списка голосований в канале)
    polls:create_index('channel', {
        type = 'TREE',
        unique = false,
        parts = {'channel_id'},
        if_not_exists = true
    })

    -- По создателю (для поиска своих голосований)
    polls:create_index('creator', {
        type = 'TREE',
        unique = false,
        parts = {'created_by'},
        if_not_exists = true
    })

    local votes = box.schema.space.create('votes', {
        if_not_exists = false,
        format = {
            {name = 'id', type = 'string'},           -- ID голоса
            {name = 'poll_id', type = 'string'},      -- ID голосования
            {name = 'user_id', type = 'string'},      -- ID пользователя
            {name = 'option_idx', type = 'number'},   -- Индекс выбранного варианта
            {name = 'created_at', type = 'number'}    -- Unix timestamp создания
        }
    })

    -- По ID голоса (первичный)
    votes:create_index('primary', {
        type = 'HASH',
        unique = true,
        parts = {'id'},
        if_not_exists = true
    })

    -- По ID голосования (для подсчета всех голосов в опросе)
    votes:create_index('poll_id', {
        type = 'TREE',
        unique = false,
        parts = {'poll_id'},
        if_not_exists = true
    })

    -- По комбинации user_id + poll_id (уникальный, предотвращает повторное голосование)
    votes:create_index('user_poll', {
        type = 'TREE',
        parts = {'user_id', 'poll_id'},
        if_not_exists = true,
        unique = true
    })

    print('Spaces and indexes have been created successfully')
end

bootstrap()

box.schema.user.grant('guest', 'read,write,execute', 'universe', nil, {if_not_exists = true})

print('Tarantool initialization completed successfully')