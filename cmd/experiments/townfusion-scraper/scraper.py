import argparse
import requests
from bs4 import BeautifulSoup
from tabulate import tabulate

BASE_URL = "https://countyfusion10.kofiletech.us"
SEARCH_URL = f"{BASE_URL}/countyweb/search/searchExecute.do?assessor=false"
RESULTS_URL = f"{BASE_URL}/countyweb/search/TownFusion/docs_SearchResultList.jsp"
NEXT_PAGE_URL = f"{BASE_URL}/countyweb/search/searchResults.do"

# Hardcoded cookies (replace with actual values)
COOKIES = {
    "JSESSIONID": "018D101634D65A42BB2171EF45931374",
    "lhnStorageType": "cookie",
    "lhnJWT": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ2aXNpdG9yIiwiZG9tYWluIjoiIiwiZXhwIjoxNzI5NjIzNzQ0LCJpYXQiOjE3Mjk1MzczNDQsImlzcyI6eyJhcHAiOiJqc19zZGsiLCJjbGllbnQiOjI1NjE1LCJjbGllbnRfbGV2ZWwiOiJlbnRlcnByaXNlIiwibGhueF9mZWF0dXJlcyI6W10sInZpc2l0b3JfdHJhY2tpbmciOnRydWV9LCJqdGkiOiJkY2Y3YWU4Yi1lOTkyLTRkMGEtOTdlZi1lZTg2NjJjZmMxMzAiLCJyZXNvdXJjZSI6eyJpZCI6ImRjZjdhZThiLWU5OTItNGQwYS05N2VmLWVlODY2MmNmYzEzMC0yNTYxNS1ydnIxWjBmIiwidHlwZSI6IkVsaXhpci5MaG5EYi5Nb2RlbC5Db3JlLlZpc2l0b3IifX0.P7ae7v4DSJBr_5SAsbxisvzKD0c4ENt-vHOCKz7p8i4",
    "lhnRefresh": "63477298-4851-4198-9da5-7012c86f8eb7",
    "lhnContact": "dcf7ae8b-e992-4d0a-97ef-ee8662cfc130-25615-rvr1Z0f"
}

def perform_search(search_params):
    data = {
        "searchCategory": "ADVANCED",
        "searchSessionId": "searchJobMain",
        "SEARCHTYPE": "allNames",
        "RECSPERPAGE": "200",
        "DATERANGE": '[{"name":"TODATE","value":"User Defined"}]',
        "PARTY": "both",
        "REMOVECHARACTERS": "true",
        **search_params
    }
    
    response = requests.post(SEARCH_URL, data=data, cookies=COOKIES)
    return response.cookies.get("JSESSIONID")

def get_results(search_session_id, page=1):
    params = {
        "scrollPos": 0,
        "searchSessionId": search_session_id,
        "resultPageAction": "nav",
        "sortColumn": "RecordDate",
        "sortDirection": "asc",
        "navDirection": "next",
        "startCursor": (page - 1) * 200,
        "pageNumber": page,
    }
    
    response = requests.get(RESULTS_URL, params=params, cookies=COOKIES)
    return response.text

def parse_results(html):
    soup = BeautifulSoup(html, 'html.parser')
    tables = soup.find_all('table')
    
    all_headers = []
    all_rows = []
    
    for table in tables:
        headers = [th.text.strip() for th in table.find_all('th')]
        rows = []
        for tr in table.find_all('tr')[1:]:  # Skip header row
            row = [td.text.strip() for td in tr.find_all('td')]
            rows.append(row)
        
        all_headers.append(headers)
        all_rows.append(rows)
    
    return all_headers, all_rows
def main():
    parser = argparse.ArgumentParser(description="Scrape Providence Recorder of Deeds")
    parser.add_argument("--name-tree", help="Name tree for search")
    parser.add_argument("--from-date", help="From date for search", default=None)
    parser.add_argument("--to-date", help="To date for search", default=None)
    parser.add_argument("--page", type=int, default=1, help="Page number")
    parser.add_argument("--input-file", help="Path to input file containing search results")
    parser.add_argument("--output-file", help="Path to save search results")
    args = parser.parse_args()

    if args.input_file:
        with open(args.input_file, 'r') as f:
            html_results = f.read()
    else:
        search_params = {}
        if args.name_tree:
            search_params["DATA_INDEX_FIELD03"] = f"NAMETREE:{args.name_tree}"
        if args.from_date:
            search_params["FROMDATE"] = args.from_date
        if args.to_date:
            search_params["TODATE"] = args.to_date

        if search_params:
            search_session_id = perform_search(search_params)
        else:
            search_session_id = "searchJobMain"  # Default value

        html_results = get_results(search_session_id, args.page)

        if args.output_file:
            with open(args.output_file, 'w') as f:
                f.write(html_results)

    headers, rows = parse_results(html_results)

    if rows:
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    else:
        print("No results found.")
if __name__ == "__main__":
    main()
