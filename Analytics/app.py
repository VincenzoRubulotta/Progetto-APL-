import matplotlib
matplotlib.use('Agg')

from flask import Flask, send_file
import pandas as pd
import matplotlib.pyplot as plt
import io
import os

app = Flask(__name__)

CSV_PATH = "/data/match_history.csv"

@app.route('/stats')
def genera_dashboard():
    if not os.path.exists(CSV_PATH):
        return "<h1>nessuna statistica disponibile!</h1><p>"
    
    try:
        df = pd.read_csv(CSV_PATH)
    except Exception as e:
        return f"<h1>Errore di lettura:</h1><p>{e}</p>"
    
    if df.empty:
        return "<h1>Il file delle statistiche è vuoto.</h1>"
    
    df.columns = df.columns.str.strip()
    
    df['Vincitore'] = df['Vincitore'].astype(str).str.strip().str.upper()
    user_wins = len(df[df['Vincitore']=='USER'])
    cpu_wins = len(df[df['Vincitore']=='CPU'])

    fig,axs = plt.subplots(1,2, figsize=(12,5))
    fig.suptitle('Dashboard Statistiche - Scopone Scientifico', fontsize=16)

    if user_wins == 0 and cpu_wins == 0:
        axs[0].text(0.5, 0.5, "Nessuna vittoria registrata", horizontalalignment='center')
        axs[0].set_axis_off()
    else:
        axs[0].pie([user_wins,cpu_wins], labels = ['Utente', 'CPU'], autopct ='%1.1f%%', colors=['#4CAF50', '#F44336'])
        axs[0].set_title('Percentuale di Vittorie Totali')

    if 'PuntiUser' in df.columns and 'PuntiCPU' in df.columns and not df.empty:
        axs[1].plot(df.index +1,df['PuntiUser'], label = 'TuoiPunti', marker='o', color = 'green')
        axs[1].plot(df.index + 1, df['PuntiCPU'], label='Punti CPU', marker='x', color='red')
        axs[1].set_xlabel('Numero Partita')
        axs[1].set_ylabel('Punti')
        axs[1].grid(True, linestyle='--', alpha=0.7)
        axs[1].legend()
    else: 
        col_names = ", ".join(list(df.columns))
        axs[1].text(0.5, 0.5, f"Colonne non trovate.\nHo trovato queste:\n{col_names}", 
                    horizontalalignment='center', color='red', wrap=True)
    axs[1].set_title('StoricoPunteggi')

    plt.tight_layout()

    img = io.BytesIO()
    plt.savefig(img,format='png')
    img.seek(0)
    plt.close()

    return send_file(img, mimetype='image/png')



if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug = True)