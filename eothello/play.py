import requests
import json
import re
import time
import subprocess
import threading
import os
import sys
import select
from bs4 import BeautifulSoup
from datetime import datetime
import logging
from dotenv import load_dotenv

# Charger les variables d'environnement depuis le fichier .env
load_dotenv(".env")

# Configuration depuis les variables d'environnement
BINARY_PATH = os.getenv("BINARY_PATH")
GAMES_CHECK_INTERVAL = int(os.getenv("GAMES_CHECK_INTERVAL", "600"))
MOVES_CHECK_INTERVAL = int(os.getenv("MOVES_CHECK_INTERVAL", "60"))
REQUEST_DELAY = int(os.getenv("REQUEST_DELAY", "1"))
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
AUTH_COOKIE = os.getenv("AUTH_COOKIE")

# Configuration du logging
log_level = getattr(logging, LOG_LEVEL.upper(), logging.INFO)
logging.basicConfig(level=log_level, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class ProcessHandler:
    """Classe similaire à pwntools pour gérer les processus avec des pipes robustes"""
    
    def __init__(self, binary_path, timeout=5.0):
        self.binary_path = binary_path
        self.timeout = timeout
        self.process = None
        self.is_alive = False
        
    def start(self):
        """Démarre le processus"""
        try:
            self.process = subprocess.Popen(
                [self.binary_path],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=0,  # Pas de buffer
                universal_newlines=True
            )
            self.is_alive = True
            logger.info(f"Processus démarré : {self.binary_path}")
            return True
        except Exception as e:
            logger.error(f"Erreur lors du démarrage du processus : {e}")
            self.is_alive = False
            return False
    
    def send(self, data):
        """Envoie des données au processus"""
        if not self.is_alive or not self.process:
            raise RuntimeError("Le processus n'est pas démarré")
        
        try:
            if not data.endswith('\n'):
                data += '\n'
            
            self.process.stdin.write(data)
            self.process.stdin.flush()
            logger.debug(f"Données envoyées : {data.strip()}")
            
        except (BrokenPipeError, OSError) as e:
            logger.error(f"Erreur lors de l'envoi : {e}")
            self.is_alive = False
            raise
    
    def recv(self, timeout=None):
        """Reçoit des données du processus avec timeout"""
        if not self.is_alive or not self.process:
            raise RuntimeError("Le processus n'est pas démarré")
        
        if timeout is None:
            timeout = self.timeout
        
        try:
            if sys.platform == "win32":
                # Sur Windows, utiliser une approche simple avec readline et threading
                import queue
                import threading
                
                result_queue = queue.Queue()
                
                def read_line():
                    try:
                        line = self.process.stdout.readline()
                        result_queue.put(('success', line))
                    except Exception as e:
                        result_queue.put(('error', str(e)))
                
                # Démarrer le thread de lecture
                read_thread = threading.Thread(target=read_line)
                read_thread.daemon = True
                read_thread.start()
                
                # Attendre le résultat avec timeout
                try:
                    status, data = result_queue.get(timeout=timeout)
                    if status == 'success':
                        if data:
                            logger.debug(f"Données reçues : {data.strip()}")
                            return data.strip()
                        else:
                            raise RuntimeError("Le processus s'est fermé")
                    else:
                        raise RuntimeError(f"Erreur de lecture : {data}")
                except queue.Empty:
                    raise TimeoutError("Timeout lors de la réception")
                    
            else:
                # Sur Unix, utiliser select
                ready, _, _ = select.select([self.process.stdout], [], [], timeout)
                if ready:
                    data = self.process.stdout.readline()
                    if data:
                        logger.debug(f"Données reçues : {data.strip()}")
                        return data.strip()
                    else:
                        raise RuntimeError("Le processus s'est fermé")
                else:
                    raise TimeoutError("Timeout lors de la réception")
                    
        except Exception as e:
            logger.error(f"Erreur lors de la réception : {e}")
            if isinstance(e, (BrokenPipeError, OSError)):
                self.is_alive = False
            raise
    
    def sendline(self, data):
        """Envoie une ligne de données"""
        self.send(data + '\n' if not data.endswith('\n') else data)
    
    def recvline(self, timeout=None):
        """Reçoit une ligne de données"""
        return self.recv(timeout)
    
    def interactive(self, data):
        """Envoie des données et attend une réponse"""
        self.send(data)
        return self.recv()
    
    def check_alive(self):
        """Vérifie si le processus est toujours vivant"""
        if self.process:
            poll_result = self.process.poll()
            if poll_result is not None:
                self.is_alive = False
                logger.warning(f"Le processus s'est arrêté avec le code : {poll_result}")
            return poll_result is None
        return False
    
    def kill(self):
        """Tue le processus"""
        if self.process:
            try:
                self.process.terminate()
                self.process.wait(timeout=3)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
            finally:
                self.is_alive = False
                logger.info("Processus arrêté")
    
    def restart(self):
        """Redémarre le processus"""
        logger.info("Redémarrage du processus...")
        self.kill()
        time.sleep(1)  # Petit délai avant redémarrage
        return self.start()
    
    def __enter__(self):
        self.start()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.kill()

class EothelloBot:
    def __init__(self, binary_path=None):
        self.base_url = "https://www.eothello.com"
        self.session = requests.Session()
        self.binary_path = binary_path
        self.engine_handler = None
        # Configuration des cookies d'authentification
        self.cookies = {
            'authentication': AUTH_COOKIE,
        }
        
        # Headers pour les requêtes
        self.headers = {
            'accept': '*/*',
            'accept-language': 'fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7',
            'cache-control': 'no-cache',
            'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
            'origin': 'https://www.eothello.com',
            'pragma': 'no-cache',
            'priority': 'u=1, i',
            'sec-ch-ua': '"Chromium";v="134", "Not:A-Brand";v="24", "Opera GX";v="119"',
            'sec-ch-ua-mobile': '?0',
            'sec-ch-ua-platform': '"Windows"',
            'sec-fetch-dest': 'empty',
            'sec-fetch-mode': 'cors',
            'sec-fetch-site': 'same-origin',
            'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 OPR/119.0.0.0',
            'x-requested-with': 'XMLHttpRequest'
        }
        
        # Liste des parties en cours {game_id: {'color': 'black'/'white', 'last_move_count': int}}
        self.current_games = {}
        
        # Initialiser la session avec les cookies
        self.session.cookies.update(self.cookies)
        
    def start_engine(self):
        """Démarre le moteur d'IA en tant que processus séparé"""
        try:
            self.engine_handler = ProcessHandler(self.binary_path, timeout=10.0)
            if self.engine_handler.start():
                logger.info(f"Moteur d'IA démarré : {self.binary_path}")
                return True
            else:
                return False
        except Exception as e:
            logger.error(f"Erreur lors du démarrage du moteur : {e}")
            return False
    
    def get_ai_move(self, position):
        """Envoie la position au moteur et récupère le coup suggéré"""
        if not self.engine_handler or not self.engine_handler.is_alive:
            logger.error("Le moteur d'IA n'est pas démarré")
            return None
            
        try:
            # Vérifier que le processus est toujours vivant
            if not self.engine_handler.check_alive():
                logger.warning("Le moteur d'IA s'est arrêté, tentative de redémarrage...")
                if not self.engine_handler.restart():
                    logger.error("Impossible de redémarrer le moteur d'IA")
                    return None
            
            # Envoyer la position au moteur et recevoir la réponse
            logger.debug(f"Envoi de la position au moteur : {position}")
            
            # Utiliser la méthode interactive pour envoyer et recevoir
            move = self.engine_handler.interactive(position)
            
            if move:
                move = move.replace('Board > ', '')
                if len(move) != 2:
                    return None
                logger.debug(f"Coup reçu du moteur : {move}")
                return move
            else:
                logger.warning("Aucun coup reçu du moteur")
                return None
                
        except TimeoutError:
            logger.error("Timeout lors de la communication avec le moteur")
            # Tenter un redémarrage en cas de timeout
            try:
                self.engine_handler.restart()
            except:
                pass
            return None
        except Exception as e:
            logger.error(f"Erreur lors de la communication avec le moteur : {e}")
            # Tenter un redémarrage en cas d'erreur critique
            try:
                self.engine_handler.restart()
            except:
                pass
            return None
    
    def fetch_current_games(self):
        """Récupère la liste des parties en cours"""
        try:
            response = self.session.get(f"{self.base_url}/get-player-current-games-list/76887/1")
            response.raise_for_status()
            
            soup = BeautifulSoup(response.json()['content'], 'html.parser')
            
            
            # Extraire les liens des parties
            game_links = soup.find_all('a', href=re.compile(r'/game/\d+'))
            
            new_games = {}
            for link in game_links:
                href = link.get('href')
                game_id_match = re.search(r'/game/(\d+)', href)
                if game_id_match:
                    game_id = game_id_match.group(1)
                    # Pour l'instant, on ne connaît pas encore la couleur
                    new_games[game_id] = {'color': None, 'last_move_count': 0}
            
            # Mettre à jour la liste des parties en cours
            removed_games = set(self.current_games.keys()) - set(new_games.keys())
            added_games = set(new_games.keys()) - set(self.current_games.keys())
            
            if added_games:
                logger.info(f"Nouvelles parties : {added_games}")
            
            # Conserver les informations existantes pour les parties qui continuent
            for game_id in new_games:
                if game_id in self.current_games:
                    new_games[game_id] = self.current_games[game_id]
            
            self.current_games = new_games
            
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des parties : {e}")
    
    def parse_game_page(self, game_id):
        """Parse une page de partie pour extraire les informations du jeu"""
        try:
            response = self.session.get(f"{self.base_url}/game/{game_id}")
            response.raise_for_status()
            
            # Chercher le script avec initializeServerGame
            script_pattern = r'server_game\.initializeServerGame\s*\(\s*([^)]+)\s*\)'
            match = re.search(script_pattern, response.text, re.DOTALL)
            
            if not match:
                logger.warning(f"Script initializeServerGame non trouvé pour la partie {game_id}")
                return None
            
            # Extraire les paramètres
            params_str = match.group(1)
            
            # Parser les paramètres (format JavaScript)
            try:
                # Nettoyer et convertir en JSON valide
                params_str = re.sub(r'(\w+):', r'"\1":', params_str)  # Ajouter des guillemets aux clés
                params_str = re.sub(r',\s*]', ']', params_str)  # Nettoyer les virgules en trop
                
                # Séparer les paramètres manuellement car c'est du JavaScript, pas du JSON
                params = []
                current_param = ""
                bracket_count = 0
                in_string = False
                escape_next = False
                
                for char in params_str:
                    if escape_next:
                        current_param += char
                        escape_next = False
                        continue
                        
                    if char == '\\':
                        escape_next = True
                        current_param += char
                        continue
                        
                    if char == '"' and not escape_next:
                        in_string = not in_string
                        current_param += char
                        continue
                        
                    if not in_string:
                        if char in '[{(':
                            bracket_count += 1
                        elif char in ']})':
                            bracket_count -= 1
                        elif char == ',' and bracket_count == 0:
                            params.append(current_param.strip())
                            current_param = ""
                            continue
                    
                    current_param += char
                
                if current_param.strip():
                    params.append(current_param.strip())
                
                # Nettoyer les paramètres
                cleaned_params = []
                for param in params:
                    param = param.strip()
                    if param.startswith('"') and param.endswith('"'):
                        cleaned_params.append(param[1:-1])  # Enlever les guillemets
                    elif param.startswith('[') and param.endswith(']'):
                        # Parser le tableau des coups
                        moves_str = param[1:-1]
                        moves = []
                        for move in re.findall(r'"([^"]+)"', moves_str):
                            moves.append(move)
                        cleaned_params.append(moves)
                    else:
                        try:
                            # Tenter de convertir en nombre
                            if '.' in param:
                                cleaned_params.append(float(param))
                            else:
                                cleaned_params.append(int(param))
                        except ValueError:
                            cleaned_params.append(param)
                
                if len(cleaned_params) >= 13:
                    game_info = {
                        'game_id': cleaned_params[0],
                        'moves': cleaned_params[1],
                        'starting_position': cleaned_params[2],
                        'winner': cleaned_params[3],
                        'variant': cleaned_params[4],
                        'game_status_text': cleaned_params[5],
                        'player_name': cleaned_params[6],
                        'role': cleaned_params[11],  # 1 pour noir, 2 pour blanc
                        'turn': cleaned_params[12]   # "black" ou "white"
                    }
                    return game_info
                else:
                    logger.error(f"Pas assez de paramètres extraits pour la partie {game_id}: {len(cleaned_params)}")
                    return None
                    
            except Exception as e:
                logger.error(f"Erreur lors du parsing des paramètres pour la partie {game_id}: {e}")
                return None
                
        except Exception as e:
            logger.error(f"Erreur lors de la récupération de la partie {game_id}: {e}")
            return None
    
    def make_move(self, game_id, move, move_index):
        """Envoie un coup au serveur"""
        try:
            data = {
                'game_id': game_id,
                'move': move,
                'move_index': move_index
            }
            
            response = self.session.post(
                f"{self.base_url}/make-move",
                headers=self.headers,
                data=data
            )
            
            response.raise_for_status()
            return True
            
        except Exception as e:
            logger.error(f"Erreur lors de l'envoi du coup {move} pour la partie {game_id}: {e}")
            return False
    
    def process_game(self, game_id):
        """Traite une partie spécifique"""
        try:
            game_info = self.parse_game_page(game_id)
            logger.debug(f"Traitement de la partie {game_id}: {game_info}")
            if not game_info:
                return
            
            # Déterminer notre couleur
            our_color = "black" if game_info['role'] == 1 else "white"
            
            # Mettre à jour les informations de la partie
            if game_id not in self.current_games:
                self.current_games[game_id] = {}
            
            self.current_games[game_id]['color'] = our_color
            current_move_count = len(game_info['moves'])
            
            # Vérifier si c'est notre tour et s'il y a de nouveaux coups
            if game_info['turn'] == our_color:
                # Construire la position actuelle
                position = "".join(game_info['moves'])
                
                # Obtenir le coup de l'IA
                ai_move = self.get_ai_move(position)
                if ai_move:
                    # Envoyer le coup
                    move_index = len(game_info['moves']) + 1
                    if self.make_move(game_id, ai_move, move_index):
                        self.current_games[game_id]['last_move_count'] = current_move_count + 1
                else:
                    logger.error(f"L'IA n'a pas pu suggérer de coup pour la partie {game_id}")
            else:
                # Juste mettre à jour le compteur de coups
                self.current_games[game_id]['last_move_count'] = current_move_count
                if game_info['turn'] != our_color:
                    logger.debug(f"Partie {game_id}: En attente du coup de l'adversaire")
                    
        except Exception as e:
            logger.error(f"Erreur lors du traitement de la partie {game_id}: {e}")
    
    def monitor_games(self):
        """Surveille toutes les parties en cours"""
          # Récupérer la liste des parties toutes les 10 minutes
        def fetch_games_periodically():
            while True:
                self.fetch_current_games()
                time.sleep(GAMES_CHECK_INTERVAL)
        
        # Démarrer le thread de récupération des parties
        games_thread = threading.Thread(target=fetch_games_periodically, daemon=True)
        games_thread.start()
        
        # Récupération initiale
        self.fetch_current_games()
        
        # Boucle principale - vérifier les parties toutes les minutes
        while True:
            try:
                if self.current_games:
                    for game_id in list(self.current_games.keys()):
                        self.process_game(game_id)
                        time.sleep(REQUEST_DELAY)  # Délai entre les parties
                
                # Attendre avant la prochaine vérification
                time.sleep(MOVES_CHECK_INTERVAL)
                
            except KeyboardInterrupt:
                logger.info("Arrêt demandé par l'utilisateur")
                break
            except Exception as e:
                logger.error(f"Erreur dans la boucle principale : {e}")
                time.sleep(30)  # Attendre 30 secondes en cas d'erreur
    
    def cleanup(self):
        """Nettoie les ressources"""
        if self.engine_handler:
            self.engine_handler.kill()
            logger.info("Moteur d'IA arrêté")

def main():
    bot = EothelloBot(BINARY_PATH)
    
    try:
        # Démarrer le moteur d'IA
        if not bot.start_engine():
            logger.error("Impossible de démarrer le moteur d'IA")
            return
        
        # Commencer la surveillance
        bot.monitor_games()
        
    except KeyboardInterrupt:
        logger.info("Arrêt du bot...")
    finally:
        bot.cleanup()

if __name__ == "__main__":
    main()
