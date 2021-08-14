#!/usr/bin/python3

import bs4
import csv
import email
import sys

out = csv.writer(sys.stdout)
out.writerow(['playerid', 'player'])

with open('draftresults.mhtml') as f:
    for part in email.message_from_file(f).walk():
        if part.get_content_type() == 'text/html':
            soup = bs4.BeautifulSoup(part.get_payload(decode=True), features='html.parser')
            break

for td in soup.find_all('td', {'class': 'player'}):
    # Only print keepers
    if td.span is not None:
        playerid = td.a['href'][29:].rstrip('/')
        player = td.a.get_text()
        out.writerow([playerid, player])
