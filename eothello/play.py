import requests
import bs4
import time
import subprocess
import logging


AUTH_COOKIE = "redacted"
ACCOUNT_ID = AUTH_COOKIE.split("%3A")[0]

# Configuration du logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('log.txt', encoding='utf-8'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

session = requests.Session()
session.cookies.set("authentication", AUTH_COOKIE, domain="www.eothello.com", path="/")

GAMES = {}

# Start cli.exe process with pipes for stdin/stdout/stderr
r = subprocess.Popen(
  ["cli.exe"],
  stdin=subprocess.PIPE,
  stdout=subprocess.PIPE,
  stderr=subprocess.PIPE,
  text=False,  # Use binary mode for communication
  bufsize=1    # Line buffering
)

# Helper methods to mimic pwntools-like interface
def recvuntil(self, delimiter):
  buffer = b""
  while not buffer.endswith(delimiter):
    char = self.stdout.read(1)
    if not char:
      break
    buffer += char
  return buffer

def recvline(self):
  return self.stdout.readline()

def sendline(self, data):
  if isinstance(data, str):
    data = data.encode()
  self.stdin.write(data + b"\n")
  self.stdin.flush()

# Add these methods to the Popen object
r.recvuntil = recvuntil.__get__(r)
r.recvline = recvline.__get__(r)
r.sendline = sendline.__get__(r)

  
def make_move(game_id, move, index):
  
  # Create form data
  data = {
    "game_id": game_id,
    "move": move,
    "move_index": index
  }
  
  res = session.post("https://www.eothello.com/make-move",
            data=data)
  if res.status_code != 200:
    logger.error("Failed to make move")
    
def collect_new_games():
    res = session.get(f"https://www.eothello.com/get-player-current-games-list/{ACCOUNT_ID}/1")
    if res.status_code == 200:
      data = res.json()["content"]
      soup = bs4.BeautifulSoup(data, "html.parser")
      games = soup.find_all("div", class_="col-7 no-decoration-links")
      for game in games:
          game_id = game.find("a")["href"].split("/")[-1]
          if game_id not in GAMES:
              fetch_game(game_id)
      
    
def fetch_game(game_id):
    res = session.get(f"https://eothello.com/game/{game_id}")
    if res.status_code == 200:
      data = res.text
      first = data.find("server_game.initializeServerGame(")
      last = data.find(");", first)
      
      arrayStart = data.find("[", first)
      arrayEnd = data.find("]", arrayStart)    
      
      params = data[arrayEnd+2:last].replace('"', '').split(",")
      role = "black" if params[9].strip() == "1" else "white"
      GAMES[game_id] = role
      logger.info(f"Playing game {game_id} as {role}")
  
def update(game_id):
    res = session.get(f"https://eothello.com/game/{game_id}")
    if res.status_code == 200:
      data = res.text
      first = data.find("server_game.initializeServerGame(")
      last = data.find(");", first)
      
      arrayStart = data.find("[", first)
      arrayEnd = data.find("]", arrayStart)    
      
      params = data[arrayEnd+2:last].replace('"', '').split(",")
      
      if params[1].strip() != "":
        logger.info(f"Game {game_id} as {GAMES[game_id]} ended, winner is {params[1].strip()}")
        del GAMES[game_id]
        return
      
      if params[10].strip() != GAMES[game_id]:
          return
        
      print(f"Time to play game {game_id} as {GAMES[game_id]}")
        
      gameData = data[arrayStart+1:arrayEnd].replace('"', '').replace(",", "").strip()
      r.recvuntil(b">")
      r.sendline(gameData)
      move = r.recvline().decode().strip()
      make_move(game_id, move, len(gameData) // 2 + 1)
      
      
if __name__ == "__main__":
  logger.info("Starting Eothello bot")
  
  collection_interval = 600
  update_interval = 60
  last_collection_time = 1
  
  collect_new_games()
  
  while True:
    for game_id in list(GAMES.keys()):
      try:
        update(game_id)
      except Exception as e:
        logger.error(f"Error updating game {game_id}: {e}")
    
    current_time = time.time()
    if current_time - last_collection_time >= collection_interval:
      try:
        collect_new_games()
        last_collection_time = current_time
      except Exception as e:
        logger.error(f"Error collecting new games: {e}")
    
    time.sleep(update_interval)
      
      
      
      