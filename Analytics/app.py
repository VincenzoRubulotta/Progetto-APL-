import matplotlib
matplotlib.use('Agg')

from flask import Flask, send_file
import pandas as pd
import matplotlib.pyplot as plt
import io
import os

app = Flask(__name__)

CSV_PATH = "/data/match_history.csv"
ROUND_CSV_PATH = "/data/rounds_history.csv"

@app.route('/stats/match/<int:game_id>')
def genera_stats_singola_partita(game_id):
    if not os.path.exists(ROUND_CSV_PATH):
        return "<h1>Nessun dato dei round!</h1><p>Finisci almeno una smazzata per vedere le statistiche.</p>"
    
    try: 
        df_rounds = pd.read_csv(ROUND_CSV_PATH)
    except Exception as e:
        return f"<h1>Errore di lettura CSV Round:</h1><p>{e}</p>"
    
    df_rounds.columns = df_rounds.columns.str.strip()
    df_match = df_rounds[df_rounds['GameID'] == game_id]
    userName = str(df_match['UserName'].iloc[0]).strip()

    if df_match.empty:
        return f"<h1>Partita {game_id} non trovata</h1><p>Nessun round registrato per questa partita completa il primo match e poi riprova.</p>"
    
    colonne_bool = colonne_bool = ['PrimieraUser', 'PrimieraCPU', 'SettebelloUser', 'SettebelloCPU', 'DenariUser', 'DenariCPU']
    for col in colonne_bool:
        df_match[col] = df_match[col].astype(str).str.lower().map({'true':1,'false':0}).fillna(0)

    fig, axs = plt.subplots(1, 3, figsize=(18, 6))
    fig.suptitle(f'Analisi Smazzate - Nome Giocatore: {userName}', fontsize=18, fontweight='bold')

    axs[0].plot(df_match['RoundNum'], df_match['CarteUser'], label='Carte Tue', marker='o', color='green', linewidth=2)
    axs[0].plot(df_match['RoundNum'], df_match['CarteCPU'], label='Carte CPU', marker='x', color='red', linewidth=2)
    axs[0].set_title('Dominio del Tavolo (Carte Prese)')
    axs[0].set_xlabel('Round')
    axs[0].set_xticks(df_match['RoundNum'])
    axs[0].grid(True, linestyle='--', alpha=0.6)
    axs[0].legend()

    width = 0.35
    axs[1].bar(df_match['RoundNum'] - width/2, df_match['ScopeUser'], width, label='Tue Scope', color='#4CAF50')
    axs[1].bar(df_match['RoundNum'] + width/2, df_match['ScopeCPU'], width, label='Scope CPU', color='#F44336')
    axs[1].set_title('Andamento Scope')
    axs[1].set_xlabel('Round')
    axs[1].set_xticks(df_match['RoundNum'])
    axs[1].yaxis.get_major_locator().set_params(integer=True) 
    axs[1].grid(axis='y', linestyle='--', alpha=0.5)
    axs[1].legend()

    tot_primiera_user = df_match['PrimieraUser'].sum()
    tot_primiera_cpu = df_match['PrimieraCPU'].sum()
    tot_settebello_user = df_match['SettebelloUser'].sum()
    tot_settebello_cpu = df_match['SettebelloCPU'].sum()
    tot_denari_user = df_match['DenariUser'].sum()
    tot_denari_cpu = df_match['DenariCPU'].sum()

    vittorie_carte_user = (df_match['CarteUser'] > df_match['CarteCPU']).sum()
    vittorie_carte_cpu = (df_match['CarteCPU'] > df_match['CarteUser']).sum()

    axs[2].axis('off')

    col_labels = ['Punti di Mazzo', 'Prese da TE\n(volte)', 'Prese da CPU\n(volte)']
    table_data = [
        ['Maggioranza Carte', f"{vittorie_carte_user}", f"{vittorie_carte_cpu}"],
        ['Carte a Denari', f"{int(tot_denari_user)}", f"{int(tot_denari_cpu)}"],
        ['Settebello', f"{int(tot_settebello_user)}", f"{int(tot_settebello_cpu)}"],
        ['Primiera', f"{int(tot_primiera_user)}", f"{int(tot_primiera_cpu)}"]
    ]

    table = axs[2].table(cellText=table_data, colLabels=col_labels, loc='center', cellLoc='center')
    table.auto_set_font_size(False)
    table.set_fontsize(13)
    table.scale(1.2, 2.5) 

    for (row, col), cell in table.get_celld().items():
        if row == 0:
            cell.set_text_props(weight='bold', color='white')
            cell.set_facecolor('#343a40')

    axs[2].set_title('Dominio Punti di Mazzo (Somma Round)', fontweight='bold', fontsize=14, pad=20)

    plt.tight_layout()
    
    img = io.BytesIO()
    plt.savefig(img, format='png')
    img.seek(0)
    plt.close()
    
    return send_file(img, mimetype='image/png')

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

    df['Scarto'] = df['PuntiUser'] - df['PuntiCPU']
    media_user = df['PuntiUser'].mean()
    media_cpu = df['PuntiCPU'].mean()
    tot_partite = len(df)
    max_scarto = df['Scarto'].max()

    fig,axs = plt.subplots(2,2, figsize=(12,10))
    fig.suptitle('Dashboard Statistiche - Scopone Scientifico', fontsize=16)

    if user_wins == 0 and cpu_wins == 0:
        axs[0,0].text(0.5, 0.5, "Nessuna vittoria registrata", horizontalalignment='center')
        axs[0,0].set_axis_off()
    else:
        axs[0,0].pie([user_wins,cpu_wins], labels = ['Utente', 'CPU'], autopct ='%1.1f%%', colors=['#4CAF50', '#F44336'])
        axs[0,0].set_title('Percentuale di Vittorie Totali')

    if 'PuntiUser' in df.columns and 'PuntiCPU' in df.columns and not df.empty:
        axs[0,1].plot(df.index +1,df['PuntiUser'], label = 'TuoiPunti', marker='o', color = 'green')
        axs[0,1].plot(df.index + 1, df['PuntiCPU'], label='Punti CPU', marker='x', color='red')
        axs[0,1].set_xlabel('Numero Partita')
        axs[0,1].set_ylabel('Punti')
        axs[0,1].grid(True, linestyle='--', alpha=0.7)
        axs[0,1].legend()
    else: 
        col_names = ", ".join(list(df.columns))
        axs[0,1].text(0.5, 0.5, f"Colonne non trovate.\nHo trovato queste:\n{col_names}", 
                    horizontalalignment='center', color='red', wrap=True)
    axs[0,1].set_title('StoricoPunteggi')

    colori_barre = ['#4CAF50' if val > 0 else '#F44336' for val in df['Scarto']]
    axs[1,0].bar(df.index +1, df['Scarto'], color = colori_barre, alpha = 0.8)
    axs[1,0].axhline(0, color = 'black', linewidth=1.5)
    axs[1,0].set_xlabel('Numero Partita')
    axs[1,0].set_ylabel('Scarto (Punti tuoi - Punti CPU)')
    axs[1,0].grid(axis='y', linestyle='--', alpha = 0.5)
    axs[1,0].set_title('Scarto Punti per Partita', fontsize=14)

    axs[1,1].axis('off')
    testo_kpi = (
        f"REPORT GLOBALE\n"
        f"----------------------------------\n"
        f"Partite Totali Giocate: {tot_partite}\n"
        f"Vittorie: {user_wins} | Sconfitte: {cpu_wins}\n\n"
        f"MEDIA PUNTI\n"
        f"----------------------------------\n"
        f"La tua media: {media_user:.1f} punti\n"
        f"Media della CPU: {media_cpu:.1f} punti\n\n"
        f"RECORD ASSOLUTI\n"
        f"----------------------------------\n"
        f"Miglior vittoria: +{int(max_scarto)} punti"
    )

    axs[1,1].text(0.1, 0.5, testo_kpi, fontsize = 15, va='center', ha='left',
                    bbox=dict(facecolor='#f8f9fa', edgecolor='#ced4da', boxstyle='round,pad=1.5'))

    plt.tight_layout()

    img = io.BytesIO()
    plt.savefig(img,format='png')
    img.seek(0)
    plt.close()

    return send_file(img, mimetype='image/png')



if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug = True)