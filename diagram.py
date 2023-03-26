# diagram.py
from diagrams import Diagram, Cluster, Edge
from diagrams.k8s.infra import ETCD
from diagrams.programming.language import Go
from diagrams.onprem.database import Couchbase
from diagrams.programming.flowchart import Action

with Diagram("search", show=False):


    with Cluster("Watcher"):
        watcher = Go("Watcher")

    with Cluster("Scanner"):
        scanner = Go("Scanner")

    with Cluster("API"):
        searchAPI= Go("Search API handler")

    etcd = ETCD("etcd")

    watcher >> Edge(color="green", style="dashed", label="watch for new keys with prefix") >> etcd
    scanner >> Edge(color="green", style="dashed", label="every 60s GET key prefix") >> etcd
    
    bleve = Couchbase("in memory Bleve index")

    watcher >> Edge(color="red", label="write index") >> bleve 
    scanner >> Edge(color="red", label="write index") >> bleve 
    
    
    Edge(label="GET /api/search?query=<x>") >> searchAPI >> Edge(color="blue", label="search index") >> bleve


