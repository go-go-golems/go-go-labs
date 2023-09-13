// For example, an object for users with a list of addresses would have the schema.
//
// tables:
//
//		  users:
//		    fields:
//		      - name: name
//		        type: string
//			    unique: true
//	 	    - name: age
//	 	      type: int
//	 	      index: true
//	 	addresses:
//	 	  fields:
//	 	    - name: street
//	 	      type: string
//	 	    - name: city
//	 	      type: string
//	 	    - name: country
//	 	      type: string
//	 	      index: true
//	 	categories:
//	 	  value-field:
//	 	    name: category
//	 	    type: string
//	 	  is-list: true
//
// main-table: users
//
// The JSON representation of a single user would be:
//
//	{
//	  "name": "John",
//	  "age": 18,
//	  "addresses": [
//	    {
//	      "street": "123 Main St",
//	      "city": "Springfield",
//	      "country": "USA"
//	    },
//	    {
//	      "street": "456 Main St",
//	      "city": "Springfield",
//	      "country": "USA"
//	    }
//	  ],
//	  "categories": ["Flowering", "Perennial"]
//	}
