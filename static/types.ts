export type WsMessage = {
    data: string,
    type: string
}

export type Message = {
    username: string,
    message: string,
}

export type User = {
    username: string,
}

export type MessageUserState = {
    username: string,
    leaved: boolean,
}
