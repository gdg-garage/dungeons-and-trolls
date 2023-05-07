const GameObjectPriorities : string[] = ["Player"]

function allChildren(go: any): any[] {
    var res: any[] = [go];
    if (go["children"] == null) {
        return res
    }
    for (const child of go["children"]) {
        res.push(...allChildren(child))
    }
    return res;
}

function foregroundGameObject(a: any, b: any) : string {
    return foregroundGameObjectType(a["type"], b["type"])
}

function foregroundGameObjectType(a: string, b: string): string {
    const idxA = GameObjectPriorities.indexOf(a);
    if (idxA == -1) { // not found in the array - should not be in the foreground
        return b
    }
    if (idxA < GameObjectPriorities.indexOf(b)) {
        return a;
    }
    return b;
}

function GameObjectToEmoji(gameObject : any) : string {
    let children = allChildren(gameObject);
    let foregroundObjectType = children[0]
    for (const c of children) {
        foregroundObjectType = foregroundGameObject(foregroundObjectType, c)
    }

    switch (foregroundObjectType) {
        case "Wall":
            return "#"
        case "Player":
            return "@"
        case "Empty":
            return "."
        case "Stairs":
            return "H"
        default:
            return "?"
    }
}

function renderMapToDom(game: any) {
    for (const [levelIdx, level] of game["map"].entries()) {
        let levelHeader = document.createElement("h2");
        levelHeader.textContent = "Level " + levelIdx;
        document.body.append(levelHeader);

        let levelMap = document.createElement("div");
        for (const row of level) {
            let mapRow = document.createElement("div");
            mapRow.className = "mapRow";
            for (const tile of row) {
                let mapTile = document.createElement("span");
                mapTile.textContent = GameObjectToEmoji(tile);
                mapTile.className = "tile";
                mapRow.append(mapTile);
            }
            levelMap.append(mapRow);
        }
        document.body.append(levelMap);
    }
}

export function renderMap() {
    fetch('http://localhost:8080')
        .then((response) => response.json())
        .then((data) => renderMapToDom(data))
}