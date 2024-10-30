// Import necessary modules
let inputData = '';

// Read entire JSON input from stdin
process.stdin.on('data', (chunk) => {
    inputData += chunk;
});

process.stdin.on('end', () => {
    try {
        // Parse the complete JSON input
        const jsonData = JSON.parse(inputData);

        // Process JSON data recursively
        const processedData = processJson(jsonData);

        // Output the processed JSON with pretty print
        console.log(JSON.stringify(processedData, null, 2));
    } catch (error) {
        console.error('Invalid JSON input:', error.message);
    }
});

// Helper function to recursively process JSON data
function processJson(obj) {
    if (Array.isArray(obj)) {
        return obj.map(processJson);
    } else if (obj !== null && typeof obj === 'object') {
        // reorder properties
        const keys  = obj.properties || obj.items ? {
            type: "",
            title: "",
            description: "",
            "x-edi": "",
            minItems: "",
            maxItems: "",
            maxLength: "",
            minLength: "",
            pattern: "",
            properties: "",
            items: "",
            edi_tag: "",
            edi_order: "",
            edi_ref: "",
            edi_type: "",
            metaInfo: "",
            isRequired: ""
        } : obj;

        let val = Object.keys(keys).reduce((acc, key) => {
            acc[key] = processJson(obj[key]);
            return acc;
        }, {});

        switch(val.edi_type) {
            case "segment":
            case "element":
            case "component":
                val["x-edi"] = {
                    "type": val.edi_type,
                    "order": val.edi_order,
                    "tag": val.edi_tag,
                    "ref": val.edi_ref,
                }
                break;
        }
        if(val.type == "array") {
            val.items["x-edi"] = {
                "type": val.edi_type,
                "order": val.edi_order,
                "tag": val.edi_tag,
                "ref": val.edi_ref,
            }
            val["x-edi"] = {
                "order": val.edi_order,
            }
        }

        delete val.edi_tag;
        delete val.edi_order;
        delete val.edi_ref;
        delete val.edi_type;
        delete val.metaInfo;
        delete val.isRequired;

        return val;
    }
    // Here you can add any custom processing logic for non-object values if needed
    return obj;
}
