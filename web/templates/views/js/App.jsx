 
class App extends React.Component {
    state = {
        selected: ""
    }

    selectHandler = (event) => {
        event.preventDefault();
        const index = document.getElementById("index").value;  
        console.log("+++ selectHandler: index:", index);
        this.setState({
            selected: index
        })
    } 
    
    searchHandler = (event) => {
        const phrase = document.getElementById("phrase").value;  
        console.log("++++ searchHandler: index: ", this.state.selected, " phrase:", phrase)
        $.get("http://localhost:3000/api/items", {index: index, phrase: phrase}, res => {
            this.setState({
            items: res
        });
    });
    }

    render() {
        return(
            
<div className="container"> 
    
    <form onSubmit={this.searchHandler}>
        <div className="row form-group">
            <label for="index" className="col-lg-2 control-label text-right">Index</label>
            <div className="col-lg-2"> 
              <select name="index" id="index" className="form-control" onChange={this.selectHandler}>
                <option value="golang" selected>Golang</option>
                <option value="gopackage">Go Package</option>  
                <option value="rust">Rust</option> 
                <option value="javascript">Javascript</option>
                <option value="solidity">Solidity</option>
                <option value="kubernetes">Kubernetes</option>  
                <option value="pdf">PDF</option> 
                <option value="web">Web</option>
                <option value="note">Notes</option> 
              </select>
              
            </div>
        </div>

        <div className="row form-group">
            <label for="phrase" className="col-lg-2 control-label text-right">Phrase</label>
            <div className="col-lg-2">              
                <input type="text" id="phrase" name="phrase" className="form-control"/>
            </div> 
        </div> 

        <div className="col-lg-3 col-lg-offset-8">  
            <button type="submit" className="btn btn-success m-4">Search Now</button> 
        </div> 

        <ItemList/>
    </form>
</div>
        )
    }
}


class ItemList extends React.Component { 
    constructor(props) {
        super(props);
        this.state = {
            items: []
        }

        this.serverRequest = this.serverRequest.bind(this);
    }
   
    serverRequest() {
        $.get("http://localhost:3000/api/items", items => {
            console.log(items);
            this.setState({
                items: items
            });
        }); 
    }

    componentDidMount() {
        this.serverRequest();
    }

    render() {
        return (
            <div className="container">  
                <div className="m-4">
                    {this.state.items.map( (item, index) => 
                       <Item key={index} item={item} />)}
                </div> 
            </div>
        ); 
    }   
}

function Item(props) {
    return(
        
        <div className="row col-lg-12">
            <div> 
                <div><b></b><b>Index:</b>{props.Index}</div>
                <div><b>Phrase:</b>{props.Phrase}</div> 
                <div><b>SOURCE:</b>{props.Source}</div>  
            </div>
            <textarea id="1" cols="120" rows="6" disabled>{props.Content}</textarea>  
        </div> 
    )
}

ReactDOM.render(<App />, document.getElementById("app"));