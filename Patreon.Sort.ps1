Add-Type -AssemblyName System.Windows.Forms

# Set the home directory for the file browser dialog
$homeDirectory = "<<< INSERT HERE >>>"

# Create a file browser dialog and set the initial directory to the home directory
$openFileDialog = New-Object System.Windows.Forms.OpenFileDialog
$openFileDialog.Filter = "CSV files (*.csv)|*.csv"
$openFileDialog.Title = "Select the input CSV file"
$openFileDialog.InitialDirectory = $homeDirectory

# Display the file browser dialog and check if a file was selected
if ($openFileDialog.ShowDialog() -eq 'OK') {
    # Get the selected file path
    $csvPath = $openFileDialog.FileName

    # Define the path to the output text file
    $textPath = "<<< INSERT HERE >>>"

    # Define the column to sort by
    $sortColumn = "Tier"
	$patronStatus = "Patron Status"

    # Import the CSV file and sort it by the specified column,
    # filtering out empty values in the "Tier" column
    $data = Import-Csv -Path $csvPath |
            #Where-Object { $_.Tier -ne "" } |
			Where-Object { $_.$patronStatus -eq "Active Patron"} |
            Sort-Object -Property $sortColumn |
            Select-Object -Property Tier, Name

    # Convert the sorted data to a string with line breaks on tier changes
    $textContent = ""
    $previousTier = $null

    foreach ($row in $data) {
        if ($row.Tier -ne $previousTier) {
            $textContent += "---------------------------`r`n"
            $previousTier = $row.Tier
        }

        $textContent += $row.Name + "`r`n"
    }

    # Export the "Name" column to the output file
    $textContent | Out-File -FilePath $textPath

    Write-Host "CSV sorted and 'Name' column exported as a text file."
} else {
    Write-Host "No file selected. Script terminated."
}
