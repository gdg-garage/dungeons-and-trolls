export function processUserInput(key: KeyboardEvent) {

    switch (key.key) {
        case "ArrowUp":
            fetch('http://localhost:8080/actions', {
                method: 'POST',
                body: JSON.stringify({ "type": "Move", "direction": "up"})
            })
            break;
        case "ArrowDown":
            fetch('http://localhost:8080/actions', {
                method: 'POST',
                body: JSON.stringify({ "type": "Move", "direction": "down"})
            })
            break;
        case "ArrowLeft":
            fetch('http://localhost:8080/actions', {
                method: 'POST',
                body: JSON.stringify({ "type": "Move", "direction": "left"})
            })
            break;
        case "ArrowRight":
            fetch('http://localhost:8080/actions', {
                method: 'POST',
                body: JSON.stringify({ "type": "Move", "direction": "right"})
            })
            break;
    }
}